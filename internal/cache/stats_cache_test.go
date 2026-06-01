package cache

import (
	"testing"
	"time"

	"github.com/lemon3studio/blkhole/internal/model"
)

// setupTestCache creates a DeviceCache with a predefined device and a new StatsCache
func setupTestCache(deviceHash string, deviceID int) StatsCache {
	dc := NewDeviceCache()
	dc.LoadDevices([]*model.Device{
		{ID: deviceID, Hash: deviceHash},
	})
	return NewStatsCache(dc)
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
		expectedBlocked    int
		expectedBlockedBks int
	}{
		{"Range24h (1-min buckets)", Range24h, 20, 4, 2, 2},
		{"Range7d (5-min buckets)", Range7d, 20, 3, 2, 1},
		{"Range30d (1-hour buckets)", Range30d, 20, 2, 2, 1},
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
		})
	}
}

func TestStatsCache_GetUserCounts(t *testing.T) {
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

	// Verify total counts across all devices
	total := 0
	for _, c := range userCounts {
		total += c.Count
	}
	if total != 12 { // 5 + 3 + 4 = 12
		t.Errorf("expected 12 total counts, got %d", total)
	}

	// Verify we have 2 distinct minute buckets
	if len(userCounts) != 2 {
		t.Errorf("expected 2 buckets, got %d", len(userCounts))
	}

	// Verify blocked total counts across all devices
	blockedTotal := 0
	for _, c := range userBlockedCounts {
		blockedTotal += c.Count
	}
	if blockedTotal != 3 { // 2 + 1 = 3
		t.Errorf("expected 3 blocked counts, got %d", blockedTotal)
	}

	// Verify we have 1 distinct minute bucket for blocked
	if len(userBlockedCounts) != 1 {
		t.Errorf("expected 1 blocked bucket, got %d", len(userBlockedCounts))
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
	deviceHash := "known-device"
	sc := setupTestCache(deviceHash, 1)

	// Since sc.Cleanup() uses time.Now(), we must base our offsets on time.Now()
	// to properly test its behavior.
	now := time.Now()

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
	// For Range24h (1-min buckets), we added at:
	// - old1Min, old5Min, old1Hour
	// - valid1Min, valid5Min, valid1Hour
	// Total 6 buckets before cleanup.
	if len(sc.GetCounts(deviceHash, Range24h)) != 6 {
		t.Errorf("Expected 6 buckets for 24h range before cleanup")
	}

	// Trigger Cleanup
	sc.Cleanup()

	// Verify old stats are removed and valid ones are kept.
	// We added old stats at 25h, 8d, 31d ago.
	// For Range24h (1-min buckets), cutoff is 24h ago.
	// So 25h, 8d, 31d will be removed.
	// The ones at <24h, <7d (e.g. 6d), <30d (e.g. 29d)
	// Actually valid5Min is 6 days ago, which is < 7d but > 24h.
	// So valid5Min will ALSO be removed from 1-min buckets!
	// Wait, let's trace this carefully:
	// old1Min (25h ago) -> removed from 1m, removed from 5m, removed from 1h
	// old5Min (8d ago) -> removed from 1m, removed from 5m, removed from 1h
	// old1Hour (31d ago) -> removed from 1m, removed from 5m, removed from 1h

	// valid1Min (23h ago) -> kept in 1m, kept in 5m, kept in 1h
	// valid5Min (6d ago) -> removed from 1m! kept in 5m, kept in 1h
	// valid1Hour (29d ago) -> removed from 1m! removed from 5m! kept in 1h

	counts24h := sc.GetCounts(deviceHash, Range24h)
	if len(counts24h) != 1 {
		t.Errorf("Expected 1 valid bucket for 24h range after cleanup, got %v", counts24h)
	} else if counts24h[0].Count != 2 {
		t.Errorf("Expected count 2, got %d", counts24h[0].Count)
	}

	// For Range7d (5-min buckets), cutoff is 7d ago.
	// old1Min (25h) -> kept! (since it's < 7d)
	// valid1Min (23h) -> kept
	// valid5Min (6d) -> kept
	// 3 buckets expected
	counts7d := sc.GetCounts(deviceHash, Range7d)
	if len(counts7d) != 3 {
		t.Errorf("Expected 3 valid buckets for 7d range after cleanup, got %v", counts7d)
	}

	// For Range30d (1-hour buckets), cutoff is 30d ago.
	// old1Min (25h) -> kept
	// old5Min (8d) -> kept
	// valid1Min (23h) -> kept
	// valid5Min (6d) -> kept
	// valid1Hour (29d) -> kept
	// 5 buckets expected
	counts30d := sc.GetCounts(deviceHash, Range30d)
	if len(counts30d) != 5 {
		t.Errorf("Expected 5 valid buckets for 30d range after cleanup, got %v", counts30d)
	}

	// Similar checks for blocked counts
	blockedCounts24h := sc.GetBlockedCounts(deviceHash, Range24h)
	if len(blockedCounts24h) != 1 {
		t.Errorf("Expected 1 valid blocked bucket for 24h range after cleanup, got %v", blockedCounts24h)
	}

	blockedCounts7d := sc.GetBlockedCounts(deviceHash, Range7d)
	if len(blockedCounts7d) != 3 {
		t.Errorf("Expected 3 valid blocked buckets for 7d range after cleanup, got %v", blockedCounts7d)
	}

	blockedCounts30d := sc.GetBlockedCounts(deviceHash, Range30d)
	if len(blockedCounts30d) != 5 {
		t.Errorf("Expected 5 valid blocked buckets for 30d range after cleanup, got %v", blockedCounts30d)
	}
}
