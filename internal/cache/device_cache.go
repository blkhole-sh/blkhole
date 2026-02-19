// Package cache provides in-memory caching for DNS blocking lookups
package cache

import "github.com/lemon3studio/blkhole/internal/model"

// DeviceCache provides fast device hash to ID lookups and device-schedule mappings
type DeviceCache interface {
	LoadDevices(devices []*model.Device)
	LoadDeviceSchedules(deviceSchedules []*model.DeviceSchedule)
	GetDeviceID(hash string) (int, bool)
	GetSchedules(deviceID int) []int
}

// deviceCache implements the DeviceCache interface
type deviceCache struct {
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
	for _, dev := range devices {
		dc.deviceHashToID[dev.Hash] = dev.ID
	}
}

// LoadDeviceSchedules populates the device-to-schedule mapping
func (dc *deviceCache) LoadDeviceSchedules(deviceSchedules []*model.DeviceSchedule) {
	for _, ds := range deviceSchedules {
		dc.deviceToSchedule[ds.DeviceID] = append(dc.deviceToSchedule[ds.DeviceID], ds.ScheduleID)
	}
}

// GetDeviceID returns the device ID for a given hash
func (dc *deviceCache) GetDeviceID(hash string) (int, bool) {
	deviceID, ok := dc.deviceHashToID[hash]
	return deviceID, ok
}

// GetSchedules returns all schedule IDs for a given device ID
func (dc *deviceCache) GetSchedules(deviceID int) []int {
	return dc.deviceToSchedule[deviceID]
}
