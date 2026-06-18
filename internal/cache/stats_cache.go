// Package cache provides in-memory caching for DNS blocking lookups
package cache

import (
	"sync"
	"time"

	"github.com/blkhole-sh/blkhole/internal/model"
)

// Time range constants
const (
	Range24h = "24h"
	Range7d  = "7d"
	Range30d = "30d"
)

// timeNow is overridden in tests so fixed-window chart buckets stay deterministic.
var timeNow = time.Now

type timeRangeConfig struct {
	step time.Duration
	span time.Duration
}

// StatsCache provides in-memory query counting per device
type StatsCache interface {
	Increment(deviceHash string)
	IncrementBlocked(deviceHash string)
	IncrementAt(deviceHash string, timestamp time.Time, count int)
	IncrementBlockedAt(deviceHash string, timestamp time.Time, count int)
	GetCounts(deviceHash string, timeRange string) []model.StatCount
	GetBlockedCounts(deviceHash string, timeRange string) []model.StatCount
	GetUserCounts(deviceHashes []string, timeRange string) []model.StatCount
	GetUserBlockedCounts(deviceHashes []string, timeRange string) []model.StatCount
	GetUserSecondCounts(deviceHashes []string) map[int64]int
	GetUserBlockedSecondCounts(deviceHashes []string) map[int64]int
	Start()
	Cleanup()
}

// bucketSeries holds per-device counts truncated to a fixed step, retained for
// the given window. One series backs one resolution (e.g. 1-minute buckets).
type bucketSeries struct {
	step   time.Duration
	retain time.Duration
	counts map[string]map[time.Time]int // deviceHash -> bucket -> count
}

func newBucketSeries(step, retain time.Duration) *bucketSeries {
	return &bucketSeries{step: step, retain: retain, counts: make(map[string]map[time.Time]int)}
}

func (b *bucketSeries) add(deviceHash string, t time.Time, count int) {
	bucket := t.Truncate(b.step)
	device := b.counts[deviceHash]
	if device == nil {
		device = make(map[time.Time]int)
		b.counts[deviceHash] = device
	}
	device[bucket] += count
}

// cleanup drops buckets older than the retention window and empty device entries.
func (b *bucketSeries) cleanup(now time.Time) {
	cutoff := now.Add(-b.retain)
	for deviceHash, counts := range b.counts {
		for timestamp := range counts {
			if timestamp.Before(cutoff) {
				delete(counts, timestamp)
			}
		}
		if len(counts) == 0 {
			delete(b.counts, deviceHash)
		}
	}
}

// statSeries bundles the three chart resolutions plus the per-second QPS samples
// for one stream of queries (total or blocked).
type statSeries struct {
	minute  *bucketSeries
	fiveMin *bucketSeries
	hour    *bucketSeries
	seconds map[string]map[int64]int // deviceHash -> unix-second -> count (24h)
}

func newStatSeries() *statSeries {
	return &statSeries{
		minute:  newBucketSeries(time.Minute, 24*time.Hour),
		fiveMin: newBucketSeries(5*time.Minute, 7*24*time.Hour),
		hour:    newBucketSeries(time.Hour, 30*24*time.Hour),
		seconds: make(map[string]map[int64]int),
	}
}

// add records count for a device at t across every resolution.
func (s *statSeries) add(deviceHash string, t time.Time, count int) {
	s.minute.add(deviceHash, t, count)
	s.fiveMin.add(deviceHash, t, count)
	s.hour.add(deviceHash, t, count)

	device := s.seconds[deviceHash]
	if device == nil {
		device = make(map[int64]int)
		s.seconds[deviceHash] = device
	}
	device[t.Unix()] += count
}

// bucketsFor returns the resolution backing the given time range, or nil.
func (s *statSeries) bucketsFor(timeRange string) *bucketSeries {
	switch timeRange {
	case Range24h:
		return s.minute
	case Range7d:
		return s.fiveMin
	case Range30d:
		return s.hour
	default:
		return nil
	}
}

func (s *statSeries) cleanup(now time.Time) {
	s.minute.cleanup(now)
	s.fiveMin.cleanup(now)
	s.hour.cleanup(now)

	// Per-second samples are kept for 24 hours.
	cutoff := now.Add(-24 * time.Hour).Unix()
	for deviceHash, counts := range s.seconds {
		for sec := range counts {
			if sec < cutoff {
				delete(counts, sec)
			}
		}
		if len(counts) == 0 {
			delete(s.seconds, deviceHash)
		}
	}
}

// statsCache implements the StatsCache interface
type statsCache struct {
	mu          sync.RWMutex
	deviceCache DeviceCache
	total       *statSeries
	blocked     *statSeries
}

// NewStatsCache creates a new stats cache instance
func NewStatsCache(deviceCache DeviceCache) StatsCache {
	return &statsCache{
		deviceCache: deviceCache,
		total:       newStatSeries(),
		blocked:     newStatSeries(),
	}
}

func getTimeRangeConfig(timeRange string) (timeRangeConfig, bool) {
	switch timeRange {
	case Range24h:
		return timeRangeConfig{step: time.Minute, span: 24 * time.Hour}, true
	case Range7d:
		return timeRangeConfig{step: 5 * time.Minute, span: 7 * 24 * time.Hour}, true
	case Range30d:
		return timeRangeConfig{step: time.Hour, span: 30 * 24 * time.Hour}, true
	default:
		return timeRangeConfig{}, false
	}
}

// fillSeries pads a fixed window so the dashboard chart keeps the full x-range.
func fillSeries(counts map[time.Time]int, timeRange string) []model.StatCount {
	config, ok := getTimeRangeConfig(timeRange)
	if !ok {
		return []model.StatCount{}
	}

	end := timeNow().Truncate(config.step).Add(config.step)
	start := end.Add(-config.span)

	result := make([]model.StatCount, 0, int(config.span/config.step))
	for timestamp := start; timestamp.Before(end); timestamp = timestamp.Add(config.step) {
		result = append(result, model.StatCount{
			Timestamp: timestamp,
			Count:     counts[timestamp],
		})
	}

	return result
}

// record adds count for a device to a series, ignoring unknown devices.
func (sc *statsCache) record(series *statSeries, deviceHash string, t time.Time, count int) {
	if _, ok := sc.deviceCache.GetDeviceID(deviceHash); !ok {
		return
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	series.add(deviceHash, t, count)
}

// Increment increments the query count for a device across all time buckets
func (sc *statsCache) Increment(deviceHash string) {
	sc.record(sc.total, deviceHash, timeNow(), 1)
}

// IncrementBlocked increments the blocked query count for a device across all time buckets
func (sc *statsCache) IncrementBlocked(deviceHash string) {
	sc.record(sc.blocked, deviceHash, timeNow(), 1)
}

// IncrementAt adds query count for a device at a specific timestamp (for testing)
func (sc *statsCache) IncrementAt(deviceHash string, timestamp time.Time, count int) {
	sc.record(sc.total, deviceHash, timestamp, count)
}

// IncrementBlockedAt adds blocked query count for a device at a specific timestamp (for testing)
func (sc *statsCache) IncrementBlockedAt(deviceHash string, timestamp time.Time, count int) {
	sc.record(sc.blocked, deviceHash, timestamp, count)
}

// deviceCounts returns the padded series for a single, known device.
func (sc *statsCache) deviceCounts(series *statSeries, deviceHash, timeRange string) []model.StatCount {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if _, ok := sc.deviceCache.GetDeviceID(deviceHash); !ok {
		return []model.StatCount{}
	}

	buckets := series.bucketsFor(timeRange)
	if buckets == nil {
		return []model.StatCount{}
	}

	return fillSeries(buckets.counts[deviceHash], timeRange)
}

// userCounts returns the padded series aggregated across the given devices.
func (sc *statsCache) userCounts(series *statSeries, deviceHashes []string, timeRange string) []model.StatCount {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	buckets := series.bucketsFor(timeRange)
	if buckets == nil {
		return []model.StatCount{}
	}

	aggregated := make(map[time.Time]int)
	for _, deviceHash := range deviceHashes {
		for timestamp, count := range buckets.counts[deviceHash] {
			aggregated[timestamp] += count
		}
	}

	return fillSeries(aggregated, timeRange)
}

// GetCounts returns query counts for a single device based on time range
func (sc *statsCache) GetCounts(deviceHash string, timeRange string) []model.StatCount {
	return sc.deviceCounts(sc.total, deviceHash, timeRange)
}

// GetBlockedCounts returns blocked query counts for a single device based on time range
func (sc *statsCache) GetBlockedCounts(deviceHash string, timeRange string) []model.StatCount {
	return sc.deviceCounts(sc.blocked, deviceHash, timeRange)
}

// GetUserCounts aggregates query counts for all user devices based on time range
func (sc *statsCache) GetUserCounts(deviceHashes []string, timeRange string) []model.StatCount {
	return sc.userCounts(sc.total, deviceHashes, timeRange)
}

// GetUserBlockedCounts aggregates blocked query counts for all user devices based on time range
func (sc *statsCache) GetUserBlockedCounts(deviceHashes []string, timeRange string) []model.StatCount {
	return sc.userCounts(sc.blocked, deviceHashes, timeRange)
}

// Start begins the background cleanup goroutine
func (sc *statsCache) Start() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			sc.Cleanup()
		}
	}()
}

// Cleanup removes old data to manage memory
// Should be called periodically (e.g., every hour)
func (sc *statsCache) Cleanup() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	now := timeNow()
	sc.total.cleanup(now)
	sc.blocked.cleanup(now)
}

// GetUserSecondCounts aggregates per-second query counts (QPS samples) across
// the given devices, keyed by unix second.
func (sc *statsCache) GetUserSecondCounts(deviceHashes []string) map[int64]int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	return aggregateSeconds(sc.total.seconds, deviceHashes)
}

// GetUserBlockedSecondCounts aggregates per-second blocked query counts across
// the given devices, keyed by unix second.
func (sc *statsCache) GetUserBlockedSecondCounts(deviceHashes []string) map[int64]int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	return aggregateSeconds(sc.blocked.seconds, deviceHashes)
}

func aggregateSeconds(secondMap map[string]map[int64]int, deviceHashes []string) map[int64]int {
	aggregated := make(map[int64]int)
	for _, deviceHash := range deviceHashes {
		for sec, count := range secondMap[deviceHash] {
			aggregated[sec] += count
		}
	}
	return aggregated
}

// qpsWindow returns the tumbling window length for the QPS series of a range:
// 5 minutes for 24h, 30 for 7d, 120 for 30d.
func qpsWindow(timeRange string) (time.Duration, bool) {
	switch timeRange {
	case Range24h:
		return 5 * time.Minute, true
	case Range7d:
		return 30 * time.Minute, true
	case Range30d:
		return 120 * time.Minute, true
	default:
		return 0, false
	}
}

// WindowQPSMaxima reduces per-second query counts to one point per tumbling
// window over the given range: the window's highest QPS sample, plotted at
// the timestamp of the second where the peak occurred. The earliest second
// wins ties; windows without queries yield zero at the window start.
func WindowQPSMaxima(seconds map[int64]int, timeRange string) []model.StatCount {
	config, ok := getTimeRangeConfig(timeRange)
	if !ok {
		return []model.StatCount{}
	}
	window, ok := qpsWindow(timeRange)
	if !ok {
		return []model.StatCount{}
	}
	windowSec := int64(window / time.Second)
	spanSec := int64(config.span / time.Second)
	end := timeNow().Unix()/windowSec*windowSec + windowSec
	start := end - spanSec

	type peak struct {
		sec   int64
		count int
	}
	peaks := make(map[int64]peak)
	for sec, count := range seconds {
		if sec < start || sec >= end {
			continue
		}
		w := sec / windowSec * windowSec
		p, ok := peaks[w]
		if !ok || count > p.count || (count == p.count && sec < p.sec) {
			peaks[w] = peak{sec: sec, count: count}
		}
	}

	result := make([]model.StatCount, 0, (end-start)/windowSec)
	for w := start; w < end; w += windowSec {
		if p, ok := peaks[w]; ok {
			result = append(result, model.StatCount{Timestamp: time.Unix(p.sec, 0).UTC(), Count: p.count})
		} else {
			result = append(result, model.StatCount{Timestamp: time.Unix(w, 0).UTC(), Count: 0})
		}
	}
	return result
}
