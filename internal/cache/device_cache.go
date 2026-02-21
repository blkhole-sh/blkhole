// Package cache provides in-memory caching for DNS blocking lookups
package cache

import (
	"sync"

	"github.com/lemon3studio/blkhole/internal/model"
)

// DeviceCache provides fast device hash to ID lookups and device-schedule mappings
type DeviceCache interface {
	LoadDevices(devices []*model.Device)
	LoadDeviceSchedules(deviceSchedules []*model.DeviceSchedule)
	GetDeviceID(hash string) (int, bool)
	GetSchedules(deviceID int) []int
}

// deviceCache implements the DeviceCache interface
type deviceCache struct {
	mu               sync.RWMutex
	deviceHashToID   map[string]int // Device hash → Device ID
	deviceToSchedule map[int][]int  // Device ID → Schedule IDs
}

// NewDeviceCache creates a new device cache instance
func NewDeviceCache() DeviceCache {
	return &deviceCache{
		deviceHashToID:   make(map[string]int),
		deviceToSchedule: make(map[int][]int),
	}
}

// LoadDevices populates the hash-to-ID mapping
func (dc *deviceCache) LoadDevices(devices []*model.Device) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	for _, dev := range devices {
		dc.deviceHashToID[dev.Hash] = dev.ID
	}
}

// LoadDeviceSchedules populates the device-to-schedule mapping
func (dc *deviceCache) LoadDeviceSchedules(deviceSchedules []*model.DeviceSchedule) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	for _, ds := range deviceSchedules {
		dc.deviceToSchedule[ds.DeviceID] = append(dc.deviceToSchedule[ds.DeviceID], ds.ScheduleID)
	}
}

// GetDeviceID returns the device ID for a given hash
func (dc *deviceCache) GetDeviceID(hash string) (int, bool) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	deviceID, ok := dc.deviceHashToID[hash]
	return deviceID, ok
}

// GetSchedules returns all schedule IDs for a given device ID
func (dc *deviceCache) GetSchedules(deviceID int) []int {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	return dc.deviceToSchedule[deviceID]
}
