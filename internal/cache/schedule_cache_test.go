package cache

import (
	"testing"

	"github.com/lemon3studio/blkhole/internal/model"
)

func TestTimeStringToSlot(t *testing.T) {
	tests := []struct {
		timeStr  string
		expected uint16
	}{
		{"00:00", 0},
		{"00:05", 1},
		{"01:00", 12},   // 1 * 12 + 0
		{"12:30", 150},  // 12 * 12 + 30/5 = 144 + 6 = 150
		{"23:55", 287},  // 23 * 12 + 55/5 = 276 + 11 = 287
		{"24:00", 0},    // invalid hour
		{"12:60", 0},    // invalid minute
		{"invalid", 0},  // invalid format
		{"12", 0},       // missing minute
	}

	for _, tc := range tests {
		t.Run(tc.timeStr, func(t *testing.T) {
			result := timeStringToSlot(tc.timeStr)
			if result != tc.expected {
				t.Errorf("timeStringToSlot(%q) = %d; want %d", tc.timeStr, result, tc.expected)
			}
		})
	}
}

func TestConvertToBitmask(t *testing.T) {
	// 1. Monday only, 00:00 - 00:05
	// Monday = day 0. 00:00 = slot 0 (bit 0), 00:05 = slot 1 (bit 0 still, since 1*8/288 = 0)
	// Actually, bits per day = 8.
	// Slot 0 to 287 are mapped to 0 to 7.
	// A slot s maps to bit (s * 8) / 288.
	// So slots 0-35 map to bit 0. Slots 36-71 map to bit 1.
	// Let's test "00:00" to "23:55" for Monday, it should set all 8 bits of Monday (day 0) = 0xFF

	schedAllDayMonday := &model.Schedule{
		Monday:    true,
		StartTime: "00:00",
		EndTime:   "23:55",
	}
	mask := convertToBitmask(schedAllDayMonday)
	if mask != 0xFF {
		t.Errorf("Expected 0xFF for all day Monday, got 0x%X", mask)
	}

	// 2. All days, 00:00 to 23:55 -> 56 bits set (8 bits per day * 7 days) -> 0x00FFFFFFFFFFFFFF
	schedAllDaysAllTimes := &model.Schedule{
		Monday:    true,
		Tuesday:   true,
		Wednesday: true,
		Thursday:  true,
		Friday:    true,
		Saturday:  true,
		Sunday:    true,
		StartTime: "00:00",
		EndTime:   "23:55",
	}
	maskAll := convertToBitmask(schedAllDaysAllTimes)
	expectedAll := uint64(0x00FFFFFFFFFFFFFF)
	if maskAll != expectedAll {
		t.Errorf("Expected 0x%X for all days all times, got 0x%X", expectedAll, maskAll)
	}

	// 3. Tuesday only, 00:00 to 02:55 -> slots 0 to 35 -> bit 0 of Tuesday
	// Tuesday is day 1. Bits 8 to 15. Bit 8 should be set.
	schedTuesdayMorning := &model.Schedule{
		Tuesday:   true,
		StartTime: "00:00",
		EndTime:   "02:55",
	}
	maskTue := convertToBitmask(schedTuesdayMorning)
	if maskTue != 0x0100 { // bit 8 set
		t.Errorf("Expected 0x0100 for Tuesday morning, got 0x%X", maskTue)
	}
}

func TestGetCurrentTimeMask(t *testing.T) {
	mask := getCurrentTimeMask()
	// Just verify it's a power of 2 (only one bit set) and non-zero
	if mask == 0 {
		t.Errorf("getCurrentTimeMask() returned 0")
	}
	// Check if only one bit is set
	if (mask & (mask - 1)) != 0 {
		t.Errorf("getCurrentTimeMask() returned more than one bit set: 0x%X", mask)
	}
}

func TestScheduleCache_LoadAndGetRules(t *testing.T) {
	sc := NewScheduleCache()
	rules := []*model.ScheduleRule{
		{ScheduleID: 1, RuleID: 100},
		{ScheduleID: 1, RuleID: 101},
		{ScheduleID: 2, RuleID: 200},
	}

	sc.LoadScheduleRules(rules)

	got1 := sc.GetRules(1)
	if len(got1) != 2 || got1[0] != 100 || got1[1] != 101 {
		t.Errorf("GetRules(1) = %v; want [100 101]", got1)
	}

	got2 := sc.GetRules(2)
	if len(got2) != 1 || got2[0] != 200 {
		t.Errorf("GetRules(2) = %v; want [200]", got2)
	}

	got3 := sc.GetRules(3)
	if len(got3) != 0 {
		t.Errorf("GetRules(3) = %v; want empty", got3)
	}
}

func TestScheduleCache_HasRuleIntersection(t *testing.T) {
	sc := NewScheduleCache()
	rules := []*model.ScheduleRule{
		{ScheduleID: 1, RuleID: 10},
		{ScheduleID: 1, RuleID: 20},
		{ScheduleID: 2, RuleID: 30},
	}
	sc.LoadScheduleRules(rules)

	tests := []struct {
		name        string
		scheduleIDs []int
		domainRules []int
		want        bool
	}{
		{"Empty inputs", []int{}, []int{}, false},
		{"No schedule IDs", []int{}, []int{10}, false},
		{"No domain rules", []int{1}, []int{}, false},
		{"Intersection with schedule 1", []int{1}, []int{20, 99}, true},
		{"No intersection with schedule 1", []int{1}, []int{30, 99}, false},
		{"Intersection with multiple schedules", []int{1, 2}, []int{99, 30}, true},
		{"Unknown schedule ID", []int{99}, []int{10}, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := sc.HasRuleIntersection(tc.scheduleIDs, tc.domainRules)
			if got != tc.want {
				t.Errorf("HasRuleIntersection(%v, %v) = %v; want %v", tc.scheduleIDs, tc.domainRules, got, tc.want)
			}
		})
	}
}

func TestScheduleCache_FilterActiveSchedules(t *testing.T) {
	sc := NewScheduleCache()

	// An always-active schedule
	s1 := &model.Schedule{
		ID:        1,
		Monday:    true,
		Tuesday:   true,
		Wednesday: true,
		Thursday:  true,
		Friday:    true,
		Saturday:  true,
		Sunday:    true,
		StartTime: "00:00",
		EndTime:   "23:55",
	}

	// A never-active schedule (no days)
	s2 := &model.Schedule{
		ID:        2,
		StartTime: "00:00",
		EndTime:   "23:55",
	}

	// A never-active schedule (no time slots, invalid format basically or 00:00 to 00:00 which could only be active for 5 mins if current time is exactly 00:00-00:05, but to be safe let's use invalid times)
	s3 := &model.Schedule{
		ID:        3,
		Monday:    true,
		Tuesday:   true,
		Wednesday: true,
		Thursday:  true,
		Friday:    true,
		Saturday:  true,
		Sunday:    true,
		StartTime: "12:60", // invalid
		EndTime:   "13:60", // invalid
	}

	sc.LoadSchedules([]*model.Schedule{s1, s2, s3})

	// To avoid flaky tests with current time, s1 is guaranteed active (all bits set),
	// s2 is guaranteed inactive (0 bits set), s3 is guaranteed inactive (0 bits set).

	active := sc.FilterActiveSchedules([]int{1, 2, 3, 4})
	if len(active) != 1 || active[0] != 1 {
		t.Errorf("FilterActiveSchedules([1, 2, 3, 4]) = %v; want [1]", active)
	}

	empty := sc.FilterActiveSchedules([]int{})
	if len(empty) != 0 {
		t.Errorf("FilterActiveSchedules([]) = %v; want []", empty)
	}
}
