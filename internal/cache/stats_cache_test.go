package cache

import (
	"testing"
	"time"

	"github.com/blkhole-sh/blkhole/internal/model"
)

// setupTestCache creates a DeviceCache with a predefined device and a new StatsCache
func setupTestCache(deviceHash string, deviceID int) StatsCache {
	dc := NewDeviceCache()
	dc.LoadDevices([]*model.Device{
		{ID: deviceID, Hash: deviceHash},
	})
	return NewStatsCache(dc)
}

func setTestNow(t *testing.T, now time.Time) {
	original := timeNow
	timeNow = func() time.Time { return now }
	t.Cleanup(func() {
		timeNow = original
	})
}

func bucketCount(timeRange string) int {
	config, ok := getTimeRangeConfig(timeRange)
	if !ok {
		return 0
	}
	return int(config.span / config.step)
}

func nonZeroBuckets(counts []model.StatCount) int {
	total := 0
	for _, count := range counts {
		if count.Count != 0 {
			total++
		}
	}
	return total
}

func countAt(counts []model.StatCount, timestamp time.Time) (int, bool) {
	for _, count := range counts {
		if count.Timestamp.Equal(timestamp) {
			return count.Count, true
		}
	}
	return 0, false
}

// TestStatsCache_IgnoreUnknownDevice verifies that unknown devices are NOT accepted.
func TestStatsCache_IgnoreUnknownDevice(t *testing.T) {
	// Create StatsCache with empty DeviceCache
	dc := NewDeviceCache()
	sc := NewStatsCache(dc)

	deviceHash := "unknown-hash"

	// Increment stats for unknown device
	sc.Increment(deviceHash)

	// Verify that stats were NOT recorded
	counts := sc.GetCounts(deviceHash, Range24h)
	if len(counts) != 0 {
		t.Errorf("StatsCache recorded stats for unknown device (vulnerability still exists)")
	}
}

// setupTestCache is already defined above...

func TestStatsCache_IncrementAndGetCounts(t *testing.T) {
	setTestNow(t, time.Date(2023, 10, 10, 13, 0, 0, 0, time.UTC))

	deviceHash := "known-device"
	sc := setupTestCache(deviceHash, 1)

	// Use explicit times to ensure predictable bucketing
	// t1 is exactly on the hour, 5-min, and 1-min boundary
	t1 := time.Date(2023, 10, 10, 12, 0, 0, 0, time.UTC)
	// t2 is +1 minute (same 5-min bucket, same hour bucket, different 1-min bucket)
	t2 := t1.Add(1 * time.Minute)
	// t3 is +5 minutes (different 5-min bucket, same hour bucket)
	t3 := t1.Add(5 * time.Minute)
	// t4 is +1 hour (different hour bucket)
	t4 := t1.Add(1 * time.Hour)

	// Increment stats at specific times
	sc.IncrementAt(deviceHash, t1, 5)
	sc.IncrementAt(deviceHash, t2, 3)
	sc.IncrementAt(deviceHash, t3, 2)
	sc.IncrementAt(deviceHash, t4, 10)

	// Increment blocked stats
	sc.IncrementBlockedAt(deviceHash, t1, 1)
	sc.IncrementBlockedAt(deviceHash, t2, 1)

	tests := []struct {
		name               string
		timeRange          string
		expectedTotal      int
		expectedBuckets    int
		expectedNonZero    int
		expectedBlocked    int
		expectedBlockedBks int
		expectedBlockedNZ  int
	}{
		{"Range24h (1-min buckets)", Range24h, 20, bucketCount(Range24h), 4, 2, bucketCount(Range24h), 2},
		{"Range7d (5-min buckets)", Range7d, 20, bucketCount(Range7d), 3, 2, bucketCount(Range7d), 1},
		{"Range30d (1-hour buckets)", Range30d, 20, bucketCount(Range30d), 2, 2, bucketCount(Range30d), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counts := sc.GetCounts(deviceHash, tt.timeRange)
			blockedCounts := sc.GetBlockedCounts(deviceHash, tt.timeRange)

			// Verify total count
			total := 0
			for _, c := range counts {
				total += c.Count
			}
			if total != tt.expectedTotal {
				t.Errorf("expected %d total counts, got %d", tt.expectedTotal, total)
			}

			// Verify number of buckets
			if len(counts) != tt.expectedBuckets {
				t.Errorf("expected %d buckets, got %d", tt.expectedBuckets, len(counts))
			}

			if nonZeroBuckets(counts) != tt.expectedNonZero {
				t.Errorf("expected %d non-zero buckets, got %d", tt.expectedNonZero, nonZeroBuckets(counts))
			}

			// Verify blocked total count
			blockedTotal := 0
			for _, c := range blockedCounts {
				blockedTotal += c.Count
			}
			if blockedTotal != tt.expectedBlocked {
				t.Errorf("expected %d blocked counts, got %d", tt.expectedBlocked, blockedTotal)
			}

			// Verify number of blocked buckets
			if len(blockedCounts) != tt.expectedBlockedBks {
				t.Errorf("expected %d blocked buckets, got %d", tt.expectedBlockedBks, len(blockedCounts))
			}

			if nonZeroBuckets(blockedCounts) != tt.expectedBlockedNZ {
				t.Errorf("expected %d non-zero blocked buckets, got %d", tt.expectedBlockedNZ, nonZeroBuckets(blockedCounts))
			}
		})
	}
}

func TestStatsCache_GetUserCounts(t *testing.T) {
	setTestNow(t, time.Date(2023, 10, 10, 13, 0, 0, 0, time.UTC))

	// Create multiple devices in DeviceCache
	dc := NewDeviceCache()
	dc.LoadDevices([]*model.Device{
		{ID: 1, Hash: "device1"},
		{ID: 2, Hash: "device2"},
	})
	sc := NewStatsCache(dc)

	t1 := time.Date(2023, 10, 10, 12, 0, 0, 0, time.UTC)
	t2 := t1.Add(1 * time.Minute)

	// Increment stats for device1
	sc.IncrementAt("device1", t1, 5)
	sc.IncrementBlockedAt("device1", t1, 2)

	// Increment stats for device2 at same time
	sc.IncrementAt("device2", t1, 3)
	sc.IncrementBlockedAt("device2", t1, 1)

	// Increment stats for device2 at different time
	sc.IncrementAt("device2", t2, 4)

	deviceHashes := []string{"device1", "device2"}

	// Get user counts
	userCounts := sc.GetUserCounts(deviceHashes, Range24h)
	userBlockedCounts := sc.GetUserBlockedCounts(deviceHashes, Range24h)

	if len(userCounts) != bucketCount(Range24h) {
		t.Fatalf("expected %d total buckets, got %d", bucketCount(Range24h), len(userCounts))
	}

	if len(userBlockedCounts) != bucketCount(Range24h) {
		t.Fatalf("expected %d blocked buckets, got %d", bucketCount(Range24h), len(userBlockedCounts))
	}

	// Verify total counts across all devices
	total := 0
	for _, c := range userCounts {
		total += c.Count
	}
	if total != 12 { // 5 + 3 + 4 = 12
		t.Errorf("expected 12 total counts, got %d", total)
	}

	// Verify we still only have 2 non-zero minute buckets
	if nonZeroBuckets(userCounts) != 2 {
		t.Errorf("expected 2 non-zero buckets, got %d", nonZeroBuckets(userCounts))
	}

	// Verify blocked total counts across all devices
	blockedTotal := 0
	for _, c := range userBlockedCounts {
		blockedTotal += c.Count
	}
	if blockedTotal != 3 { // 2 + 1 = 3
		t.Errorf("expected 3 blocked counts, got %d", blockedTotal)
	}

	// Verify we still only have 1 non-zero minute bucket for blocked
	if nonZeroBuckets(userBlockedCounts) != 1 {
		t.Errorf("expected 1 non-zero blocked bucket, got %d", nonZeroBuckets(userBlockedCounts))
	}

	// Specific bucket validation (t1 should have 8 counts)
	foundT1 := false
	for _, c := range userCounts {
		if c.Timestamp.Equal(t1) {
			foundT1 = true
			if c.Count != 8 {
				t.Errorf("expected 8 counts at t1, got %d", c.Count)
			}
		}
	}
	if !foundT1 {
		t.Errorf("expected to find counts for t1")
	}
}

func TestStatsCache_Cleanup(t *testing.T) {
	now := time.Date(2023, 10, 10, 12, 0, 0, 0, time.UTC)
	setTestNow(t, now)

	deviceHash := "known-device"
	sc := setupTestCache(deviceHash, 1)

	// Old timestamps beyond cleanup threshold
	old1Min := now.Add(-25 * time.Hour).Truncate(time.Minute)         // > 24h
	old5Min := now.Add(-8 * 24 * time.Hour).Truncate(5 * time.Minute) // > 7d
	old1Hour := now.Add(-31 * 24 * time.Hour).Truncate(time.Hour)     // > 30d

	// Valid timestamps within cleanup threshold
	// Note: Cleanup thresholds use time.Now() exactly.
	// So 23h, 6d, 29d.
	valid1Min := now.Add(-23 * time.Hour).Truncate(time.Minute)         // < 24h
	valid5Min := now.Add(-6 * 24 * time.Hour).Truncate(5 * time.Minute) // < 7d
	valid1Hour := now.Add(-29 * 24 * time.Hour).Truncate(time.Hour)     // < 30d

	// Add old stats
	sc.IncrementAt(deviceHash, old1Min, 1)
	sc.IncrementBlockedAt(deviceHash, old1Min, 1)
	sc.IncrementAt(deviceHash, old5Min, 1)
	sc.IncrementBlockedAt(deviceHash, old5Min, 1)
	sc.IncrementAt(deviceHash, old1Hour, 1)
	sc.IncrementBlockedAt(deviceHash, old1Hour, 1)

	// Add valid stats
	sc.IncrementAt(deviceHash, valid1Min, 2)
	sc.IncrementBlockedAt(deviceHash, valid1Min, 2)
	sc.IncrementAt(deviceHash, valid5Min, 2)
	sc.IncrementBlockedAt(deviceHash, valid5Min, 2)
	sc.IncrementAt(deviceHash, valid1Hour, 2)
	sc.IncrementBlockedAt(deviceHash, valid1Hour, 2)

	// Since IncrementAt adds to *all* buckets (1m, 5m, 1h),
	// each increment will add to all three granularities.
	// For Range24h (1-min buckets), the series is always padded to the full window.
	if len(sc.GetCounts(deviceHash, Range24h)) != bucketCount(Range24h) {
		t.Errorf("Expected full 24h window before cleanup")
	}

	// Trigger Cleanup
	sc.Cleanup()

	counts24h := sc.GetCounts(deviceHash, Range24h)
	if len(counts24h) != bucketCount(Range24h) {
		t.Fatalf("Expected full 24h range after cleanup, got %d buckets", len(counts24h))
	}
	if nonZeroBuckets(counts24h) != 1 {
		t.Errorf("Expected 1 valid 24h bucket after cleanup, got %d", nonZeroBuckets(counts24h))
	}
	if count, ok := countAt(counts24h, valid1Min); !ok || count != 2 {
		t.Errorf("Expected count 2 at valid1Min, got %d (found=%t)", count, ok)
	}

	counts7d := sc.GetCounts(deviceHash, Range7d)
	if len(counts7d) != bucketCount(Range7d) {
		t.Fatalf("Expected full 7d range after cleanup, got %d buckets", len(counts7d))
	}
	if nonZeroBuckets(counts7d) != 3 {
		t.Errorf("Expected 3 valid 7d buckets after cleanup, got %d", nonZeroBuckets(counts7d))
	}

	counts30d := sc.GetCounts(deviceHash, Range30d)
	if len(counts30d) != bucketCount(Range30d) {
		t.Fatalf("Expected full 30d range after cleanup, got %d buckets", len(counts30d))
	}
	if nonZeroBuckets(counts30d) != 5 {
		t.Errorf("Expected 5 valid 30d buckets after cleanup, got %d", nonZeroBuckets(counts30d))
	}

	blockedCounts24h := sc.GetBlockedCounts(deviceHash, Range24h)
	if len(blockedCounts24h) != bucketCount(Range24h) {
		t.Fatalf("Expected full 24h blocked range after cleanup, got %d buckets", len(blockedCounts24h))
	}
	if nonZeroBuckets(blockedCounts24h) != 1 {
		t.Errorf("Expected 1 valid blocked 24h bucket after cleanup, got %d", nonZeroBuckets(blockedCounts24h))
	}
	if count, ok := countAt(blockedCounts24h, valid1Min); !ok || count != 2 {
		t.Errorf("Expected blocked count 2 at valid1Min, got %d (found=%t)", count, ok)
	}

	blockedCounts7d := sc.GetBlockedCounts(deviceHash, Range7d)
	if len(blockedCounts7d) != bucketCount(Range7d) {
		t.Fatalf("Expected full 7d blocked range after cleanup, got %d buckets", len(blockedCounts7d))
	}
	if nonZeroBuckets(blockedCounts7d) != 3 {
		t.Errorf("Expected 3 valid blocked 7d buckets after cleanup, got %d", nonZeroBuckets(blockedCounts7d))
	}

	blockedCounts30d := sc.GetBlockedCounts(deviceHash, Range30d)
	if len(blockedCounts30d) != bucketCount(Range30d) {
		t.Fatalf("Expected full 30d blocked range after cleanup, got %d buckets", len(blockedCounts30d))
	}
	if nonZeroBuckets(blockedCounts30d) != 5 {
		t.Errorf("Expected 5 valid blocked 30d buckets after cleanup, got %d", nonZeroBuckets(blockedCounts30d))
	}
}
