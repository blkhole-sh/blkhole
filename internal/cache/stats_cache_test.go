package cache

import (
	"testing"
)

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
