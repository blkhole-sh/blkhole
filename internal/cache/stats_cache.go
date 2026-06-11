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
	Start()
	Cleanup()
}

// statsCache implements the StatsCache interface
type statsCache struct {
	mu                   sync.RWMutex
	deviceCache          DeviceCache
	minuteCounts         map[string]map[time.Time]int // deviceHash -> 1-min buckets (24h)
	fiveMinCounts        map[string]map[time.Time]int // deviceHash -> 5-min buckets (7d)
	hourCounts           map[string]map[time.Time]int // deviceHash -> 1-hour buckets (30d)
	minuteBlockedCounts  map[string]map[time.Time]int // deviceHash -> 1-min blocked buckets (24h)
	fiveMinBlockedCounts map[string]map[time.Time]int // deviceHash -> 5-min blocked buckets (7d)
	hourBlockedCounts    map[string]map[time.Time]int // deviceHash -> 1-hour blocked buckets (30d)
}

// NewStatsCache creates a new stats cache instance
func NewStatsCache(deviceCache DeviceCache) StatsCache {
	return &statsCache{
		deviceCache:          deviceCache,
		minuteCounts:         make(map[string]map[time.Time]int),
		fiveMinCounts:        make(map[string]map[time.Time]int),
		hourCounts:           make(map[string]map[time.Time]int),
		minuteBlockedCounts:  make(map[string]map[time.Time]int),
		fiveMinBlockedCounts: make(map[string]map[time.Time]int),
		hourBlockedCounts:    make(map[string]map[time.Time]int),
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

// Increment increments the query count for a device across all time buckets
func (sc *statsCache) Increment(deviceHash string) {
	// Validate device exists before caching stats
	if _, ok := sc.deviceCache.GetDeviceID(deviceHash); !ok {
		return
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	now := timeNow()

	// Increment 1-minute bucket
	minuteBucket := now.Truncate(time.Minute)
	if sc.minuteCounts[deviceHash] == nil {
		sc.minuteCounts[deviceHash] = make(map[time.Time]int)
	}
	sc.minuteCounts[deviceHash][minuteBucket]++

	// Increment 5-minute bucket
	fiveMinBucket := now.Truncate(5 * time.Minute)
	if sc.fiveMinCounts[deviceHash] == nil {
		sc.fiveMinCounts[deviceHash] = make(map[time.Time]int)
	}
	sc.fiveMinCounts[deviceHash][fiveMinBucket]++

	// Increment 1-hour bucket
	hourBucket := now.Truncate(time.Hour)
	if sc.hourCounts[deviceHash] == nil {
		sc.hourCounts[deviceHash] = make(map[time.Time]int)
	}
	sc.hourCounts[deviceHash][hourBucket]++
}

// IncrementBlocked increments the blocked query count for a device across all time buckets
func (sc *statsCache) IncrementBlocked(deviceHash string) {
	// Validate device exists before caching stats
	if _, ok := sc.deviceCache.GetDeviceID(deviceHash); !ok {
		return
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	now := timeNow()

	// Increment 1-minute blocked bucket
	minuteBucket := now.Truncate(time.Minute)
	if sc.minuteBlockedCounts[deviceHash] == nil {
		sc.minuteBlockedCounts[deviceHash] = make(map[time.Time]int)
	}
	sc.minuteBlockedCounts[deviceHash][minuteBucket]++

	// Increment 5-minute blocked bucket
	fiveMinBucket := now.Truncate(5 * time.Minute)
	if sc.fiveMinBlockedCounts[deviceHash] == nil {
		sc.fiveMinBlockedCounts[deviceHash] = make(map[time.Time]int)
	}
	sc.fiveMinBlockedCounts[deviceHash][fiveMinBucket]++

	// Increment 1-hour blocked bucket
	hourBucket := now.Truncate(time.Hour)
	if sc.hourBlockedCounts[deviceHash] == nil {
		sc.hourBlockedCounts[deviceHash] = make(map[time.Time]int)
	}
	sc.hourBlockedCounts[deviceHash][hourBucket]++
}

// IncrementAt adds query count for a device at a specific timestamp (for testing)
func (sc *statsCache) IncrementAt(deviceHash string, timestamp time.Time, count int) {
	// Validate device exists before caching stats
	if _, ok := sc.deviceCache.GetDeviceID(deviceHash); !ok {
		return
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	// Increment 1-minute bucket
	minuteBucket := timestamp.Truncate(time.Minute)
	if sc.minuteCounts[deviceHash] == nil {
		sc.minuteCounts[deviceHash] = make(map[time.Time]int)
	}
	sc.minuteCounts[deviceHash][minuteBucket] += count

	// Increment 5-minute bucket
	fiveMinBucket := timestamp.Truncate(5 * time.Minute)
	if sc.fiveMinCounts[deviceHash] == nil {
		sc.fiveMinCounts[deviceHash] = make(map[time.Time]int)
	}
	sc.fiveMinCounts[deviceHash][fiveMinBucket] += count

	// Increment 1-hour bucket
	hourBucket := timestamp.Truncate(time.Hour)
	if sc.hourCounts[deviceHash] == nil {
		sc.hourCounts[deviceHash] = make(map[time.Time]int)
	}
	sc.hourCounts[deviceHash][hourBucket] += count
}

// IncrementBlockedAt adds blocked query count for a device at a specific timestamp (for testing)
func (sc *statsCache) IncrementBlockedAt(deviceHash string, timestamp time.Time, count int) {
	// Validate device exists before caching stats
	if _, ok := sc.deviceCache.GetDeviceID(deviceHash); !ok {
		return
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	// Increment 1-minute blocked bucket
	minuteBucket := timestamp.Truncate(time.Minute)
	if sc.minuteBlockedCounts[deviceHash] == nil {
		sc.minuteBlockedCounts[deviceHash] = make(map[time.Time]int)
	}
	sc.minuteBlockedCounts[deviceHash][minuteBucket] += count

	// Increment 5-minute blocked bucket
	fiveMinBucket := timestamp.Truncate(5 * time.Minute)
	if sc.fiveMinBlockedCounts[deviceHash] == nil {
		sc.fiveMinBlockedCounts[deviceHash] = make(map[time.Time]int)
	}
	sc.fiveMinBlockedCounts[deviceHash][fiveMinBucket] += count

	// Increment 1-hour blocked bucket
	hourBucket := timestamp.Truncate(time.Hour)
	if sc.hourBlockedCounts[deviceHash] == nil {
		sc.hourBlockedCounts[deviceHash] = make(map[time.Time]int)
	}
	sc.hourBlockedCounts[deviceHash][hourBucket] += count
}

// selectBucket returns the appropriate bucket map based on time range
func (sc *statsCache) selectBucket(timeRange string) map[string]map[time.Time]int {
	switch timeRange {
	case Range24h:
		return sc.minuteCounts
	case Range7d:
		return sc.fiveMinCounts
	case Range30d:
		return sc.hourCounts
	default:
		return nil
	}
}

// selectBlockedBucket returns the appropriate blocked bucket map based on time range
func (sc *statsCache) selectBlockedBucket(timeRange string) map[string]map[time.Time]int {
	switch timeRange {
	case Range24h:
		return sc.minuteBlockedCounts
	case Range7d:
		return sc.fiveMinBlockedCounts
	case Range30d:
		return sc.hourBlockedCounts
	default:
		return nil
	}
}

// GetCounts returns query counts for a single device based on time range
func (sc *statsCache) GetCounts(deviceHash string, timeRange string) []model.StatCount {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if _, ok := sc.deviceCache.GetDeviceID(deviceHash); !ok {
		return []model.StatCount{}
	}

	bucketMap := sc.selectBucket(timeRange)
	if bucketMap == nil {
		return []model.StatCount{}
	}

	return fillSeries(bucketMap[deviceHash], timeRange)
}

// GetUserCounts aggregates query counts for all user devices based on time range
func (sc *statsCache) GetUserCounts(deviceHashes []string, timeRange string) []model.StatCount {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	bucketMap := sc.selectBucket(timeRange)
	if bucketMap == nil {
		return []model.StatCount{}
	}

	// Aggregate counts from all devices by timestamp
	aggregated := make(map[time.Time]int)
	for _, deviceHash := range deviceHashes {
		deviceCounts := bucketMap[deviceHash]
		if deviceCounts == nil {
			continue
		}

		for timestamp, count := range deviceCounts {
			aggregated[timestamp] += count
		}
	}

	return fillSeries(aggregated, timeRange)
}

// GetBlockedCounts returns blocked query counts for a single device based on time range
func (sc *statsCache) GetBlockedCounts(deviceHash string, timeRange string) []model.StatCount {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if _, ok := sc.deviceCache.GetDeviceID(deviceHash); !ok {
		return []model.StatCount{}
	}

	bucketMap := sc.selectBlockedBucket(timeRange)
	if bucketMap == nil {
		return []model.StatCount{}
	}

	return fillSeries(bucketMap[deviceHash], timeRange)
}

// GetUserBlockedCounts aggregates blocked query counts for all user devices based on time range
func (sc *statsCache) GetUserBlockedCounts(deviceHashes []string, timeRange string) []model.StatCount {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	bucketMap := sc.selectBlockedBucket(timeRange)
	if bucketMap == nil {
		return []model.StatCount{}
	}

	// Aggregate counts from all devices by timestamp
	aggregated := make(map[time.Time]int)
	for _, deviceHash := range deviceHashes {
		deviceCounts := bucketMap[deviceHash]
		if deviceCounts == nil {
			continue
		}

		for timestamp, count := range deviceCounts {
			aggregated[timestamp] += count
		}
	}

	return fillSeries(aggregated, timeRange)
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

	// Clean minuteCounts: remove entries older than 24 hours
	cutoff24h := now.Add(-24 * time.Hour)
	for deviceHash, counts := range sc.minuteCounts {
		for timestamp := range counts {
			if timestamp.Before(cutoff24h) {
				delete(counts, timestamp)
			}
		}
		// Remove device entry if no counts left
		if len(counts) == 0 {
			delete(sc.minuteCounts, deviceHash)
		}
	}

	// Clean minuteBlockedCounts: remove entries older than 24 hours
	for deviceHash, counts := range sc.minuteBlockedCounts {
		for timestamp := range counts {
			if timestamp.Before(cutoff24h) {
				delete(counts, timestamp)
			}
		}
		if len(counts) == 0 {
			delete(sc.minuteBlockedCounts, deviceHash)
		}
	}

	// Clean fiveMinCounts: remove entries older than 7 days
	cutoff7d := now.Add(-7 * 24 * time.Hour)
	for deviceHash, counts := range sc.fiveMinCounts {
		for timestamp := range counts {
			if timestamp.Before(cutoff7d) {
				delete(counts, timestamp)
			}
		}
		if len(counts) == 0 {
			delete(sc.fiveMinCounts, deviceHash)
		}
	}

	// Clean fiveMinBlockedCounts: remove entries older than 7 days
	for deviceHash, counts := range sc.fiveMinBlockedCounts {
		for timestamp := range counts {
			if timestamp.Before(cutoff7d) {
				delete(counts, timestamp)
			}
		}
		if len(counts) == 0 {
			delete(sc.fiveMinBlockedCounts, deviceHash)
		}
	}

	// Clean hourCounts: remove entries older than 30 days
	cutoff30d := now.Add(-30 * 24 * time.Hour)
	for deviceHash, counts := range sc.hourCounts {
		for timestamp := range counts {
			if timestamp.Before(cutoff30d) {
				delete(counts, timestamp)
			}
		}
		if len(counts) == 0 {
			delete(sc.hourCounts, deviceHash)
		}
	}

	// Clean hourBlockedCounts: remove entries older than 30 days
	for deviceHash, counts := range sc.hourBlockedCounts {
		for timestamp := range counts {
			if timestamp.Before(cutoff30d) {
				delete(counts, timestamp)
			}
		}
		if len(counts) == 0 {
			delete(sc.hourBlockedCounts, deviceHash)
		}
	}
}
