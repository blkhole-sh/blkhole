package cache

import (
	"testing"

	"github.com/lemon3studio/blkhole/internal/model"
)

func TestDeviceCache_GetDeviceID_Hit(t *testing.T) {
	dc := NewDeviceCache()
	dc.LoadDevices([]*model.Device{
		{ID: 1, Hash: "hash-aaa"},
		{ID: 2, Hash: "hash-bbb"},
	})

	tests := []struct {
		hash   string
		wantID int
		wantOK bool
	}{
		{"hash-aaa", 1, true},
		{"hash-bbb", 2, true},
		{"hash-unknown", 0, false},
	}

	for _, tc := range tests {
		t.Run(tc.hash, func(t *testing.T) {
			id, ok := dc.GetDeviceID(tc.hash)
			if ok != tc.wantOK {
				t.Errorf("GetDeviceID(%q) ok = %v; want %v", tc.hash, ok, tc.wantOK)
			}
			if id != tc.wantID {
				t.Errorf("GetDeviceID(%q) id = %d; want %d", tc.hash, id, tc.wantID)
			}
		})
	}
}

func TestDeviceCache_GetDeviceID_EmptyCache(t *testing.T) {
	dc := NewDeviceCache()
	id, ok := dc.GetDeviceID("nonexistent-hash")
	if ok {
		t.Errorf("expected cache miss for empty cache, got ok=true with id=%d", id)
	}
	if id != 0 {
		t.Errorf("expected id=0 on cache miss, got %d", id)
	}
}

func TestDeviceCache_LoadDevices_OverwritesEntry(t *testing.T) {
	dc := NewDeviceCache()
	dc.LoadDevices([]*model.Device{{ID: 1, Hash: "hash-x"}})
	dc.LoadDevices([]*model.Device{{ID: 99, Hash: "hash-x"}})

	id, ok := dc.GetDeviceID("hash-x")
	if !ok || id != 99 {
		t.Errorf("after re-load: GetDeviceID(%q) = (%d, %v); want (99, true)", "hash-x", id, ok)
	}
}

func TestDeviceCache_GetSchedules_Hit(t *testing.T) {
	dc := NewDeviceCache()
	dc.LoadDeviceSchedules([]*model.DeviceSchedule{
		{DeviceID: 1, ScheduleID: 10},
		{DeviceID: 1, ScheduleID: 20},
		{DeviceID: 2, ScheduleID: 30},
	})

	s1 := dc.GetSchedules(1)
	if len(s1) != 2 {
		t.Fatalf("GetSchedules(1) returned %d entries; want 2", len(s1))
	}
	if s1[0] != 10 || s1[1] != 20 {
		t.Errorf("GetSchedules(1) = %v; want [10 20]", s1)
	}

	s2 := dc.GetSchedules(2)
	if len(s2) != 1 || s2[0] != 30 {
		t.Errorf("GetSchedules(2) = %v; want [30]", s2)
	}
}

func TestDeviceCache_GetSchedules_EmptyCache(t *testing.T) {
	dc := NewDeviceCache()
	if schedules := dc.GetSchedules(1); len(schedules) != 0 {
		t.Errorf("GetSchedules on empty cache = %v; want empty", schedules)
	}
}

func TestDeviceCache_GetSchedules_UnknownDevice(t *testing.T) {
	dc := NewDeviceCache()
	dc.LoadDeviceSchedules([]*model.DeviceSchedule{{DeviceID: 1, ScheduleID: 10}})
	if schedules := dc.GetSchedules(999); len(schedules) != 0 {
		t.Errorf("GetSchedules(999) = %v; want empty", schedules)
	}
}
