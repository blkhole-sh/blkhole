// Package cache provides in-memory caching for DNS blocking lookups
package cache

import (
	"sync"
	"sync/atomic"

	"github.com/blkhole-sh/blkhole/internal/model"
)

// DeviceCache provides fast device hash to ID lookups and device-schedule mappings
type DeviceCache interface {
	Reset()
	Load(devices []*model.Device, deviceSchedules []*model.DeviceSchedule)
	LoadDevices(devices []*model.Device)
	LoadDeviceSchedules(deviceSchedules []*model.DeviceSchedule)
	GetDeviceID(hash string) (int, bool)
	GetSchedules(deviceID int) []int
}

type deviceSnapshot struct {
	deviceHashToID   map[string]int // Device hash → Device ID
	deviceToSchedule map[int][]int  // Device ID → Schedule IDs
}

// deviceCache implements the DeviceCache interface
type deviceCache struct {
	writeMu  sync.Mutex
	snapshot atomic.Pointer[deviceSnapshot]
}

// NewDeviceCache creates a new device cache instance
func NewDeviceCache() DeviceCache {
	dc := &deviceCache{}
	dc.snapshot.Store(newDeviceSnapshot())
	return dc
}

func newDeviceSnapshot() *deviceSnapshot {
	return &deviceSnapshot{
		deviceHashToID:   make(map[string]int),
		deviceToSchedule: make(map[int][]int),
	}
}

func buildDeviceHashToID(devices []*model.Device) map[string]int {
	deviceHashToID := make(map[string]int)
	for _, dev := range devices {
		deviceHashToID[dev.Hash] = dev.ID
	}
	return deviceHashToID
}

func buildDeviceSchedules(deviceSchedules []*model.DeviceSchedule) map[int][]int {
	deviceToSchedule := make(map[int][]int)
	for _, ds := range deviceSchedules {
		deviceToSchedule[ds.DeviceID] = append(deviceToSchedule[ds.DeviceID], ds.ScheduleID)
	}
	return deviceToSchedule
}

func (dc *deviceCache) current() *deviceSnapshot {
	if snapshot := dc.snapshot.Load(); snapshot != nil {
		return snapshot
	}
	return newDeviceSnapshot()
}

// Reset clears all cached data so the cache can be reloaded in-place
func (dc *deviceCache) Reset() {
	dc.writeMu.Lock()
	defer dc.writeMu.Unlock()

	dc.snapshot.Store(newDeviceSnapshot())
}

// Load publishes a complete device cache snapshot.
func (dc *deviceCache) Load(devices []*model.Device, deviceSchedules []*model.DeviceSchedule) {
	dc.writeMu.Lock()
	defer dc.writeMu.Unlock()

	dc.snapshot.Store(&deviceSnapshot{
		deviceHashToID:   buildDeviceHashToID(devices),
		deviceToSchedule: buildDeviceSchedules(deviceSchedules),
	})
}

// LoadDevices populates the hash-to-ID mapping
func (dc *deviceCache) LoadDevices(devices []*model.Device) {
	dc.writeMu.Lock()
	defer dc.writeMu.Unlock()

	current := dc.current()
	dc.snapshot.Store(&deviceSnapshot{
		deviceHashToID:   buildDeviceHashToID(devices),
		deviceToSchedule: current.deviceToSchedule,
	})
}

// LoadDeviceSchedules populates the device-to-schedule mapping
func (dc *deviceCache) LoadDeviceSchedules(deviceSchedules []*model.DeviceSchedule) {
	dc.writeMu.Lock()
	defer dc.writeMu.Unlock()

	current := dc.current()
	dc.snapshot.Store(&deviceSnapshot{
		deviceHashToID:   current.deviceHashToID,
		deviceToSchedule: buildDeviceSchedules(deviceSchedules),
	})
}

// GetDeviceID returns the device ID for a given hash
func (dc *deviceCache) GetDeviceID(hash string) (int, bool) {
	deviceID, ok := dc.current().deviceHashToID[hash]
	return deviceID, ok
}

// GetSchedules returns all schedule IDs for a given device ID
func (dc *deviceCache) GetSchedules(deviceID int) []int {
	return append([]int(nil), dc.current().deviceToSchedule[deviceID]...)
}
