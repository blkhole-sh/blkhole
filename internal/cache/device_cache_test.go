package cache

import (
	"testing"

	"github.com/lemon3studio/blkhole/internal/model"
)

func TestDeviceCache_Reset_ClearsAllEntries(t *testing.T) {
	dc := NewDeviceCache()
	dc.LoadDevices([]*model.Device{{ID: 1, Hash: "hash-a"}})
	dc.LoadDeviceSchedules([]*model.DeviceSchedule{{DeviceID: 1, ScheduleID: 10}})

	dc.Reset()

	if _, ok := dc.GetDeviceID("hash-a"); ok {
		t.Error("expected device to be absent after Reset")
	}
	if s := dc.GetSchedules(1); len(s) != 0 {
		t.Errorf("expected empty schedules after Reset, got %v", s)
	}
}

func TestDeviceCache_Reset_CanReloadAfterReset(t *testing.T) {
	dc := NewDeviceCache()
	dc.LoadDevices([]*model.Device{{ID: 1, Hash: "old-hash"}})
	dc.Reset()
	dc.LoadDevices([]*model.Device{{ID: 2, Hash: "new-hash"}})

	if _, ok := dc.GetDeviceID("old-hash"); ok {
		t.Error("old entry should be gone after Reset + reload")
	}
	id, ok := dc.GetDeviceID("new-hash")
	if !ok || id != 2 {
		t.Errorf("GetDeviceID(new-hash) = (%d, %v); want (2, true)", id, ok)
	}
}
