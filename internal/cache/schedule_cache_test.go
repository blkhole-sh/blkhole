package cache

import (
	"testing"
	"time"

	"github.com/blkhole-sh/blkhole/internal/model"
)

func TestTimeStringToSlot(t *testing.T) {
	tests := []struct {
		timeStr  string
		expected uint16
	}{
		{"00:00", 0},
		{"00:05", 1},
		{"01:00", 12},  // 1 * 12 + 0
		{"12:30", 150}, // 12 * 12 + 30/5 = 144 + 6 = 150
		{"23:55", 287}, // 23 * 12 + 55/5 = 276 + 11 = 287
		{"24:00", 0},   // invalid hour
		{"12:60", 0},   // invalid minute
		{"invalid", 0}, // invalid format
		{"12", 0},      // missing minute
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

func TestScheduleWindow_CoversSlot(t *testing.T) {
	morning := convertToWindow(&model.Schedule{
		Active:    true,
		Monday:    true,
		StartTime: "09:00",
		EndTime:   "10:00",
	})
	allDay := convertToWindow(&model.Schedule{
		Active:    true,
		Tuesday:   true,
		StartTime: "00:00",
		EndTime:   "00:00",
	})
	overnight := convertToWindow(&model.Schedule{
		Active:    true,
		Monday:    true,
		StartTime: "22:00",
		EndTime:   "06:00",
	})
	inactive := convertToWindow(&model.Schedule{
		Active:    false,
		Monday:    true,
		StartTime: "00:00",
		EndTime:   "00:00",
	})

	tests := []struct {
		name   string
		window scheduleWindow
		day    uint8
		slot   uint16
		want   bool
	}{
		{"morning at 09:00", morning, 0, timeStringToSlot("09:00"), true},
		{"morning at 10:00", morning, 0, timeStringToSlot("10:00"), true},
		{"morning at 10:05 just after end", morning, 0, timeStringToSlot("10:05"), false},
		{"morning at 08:55 just before start", morning, 0, timeStringToSlot("08:55"), false},
		{"morning at 11:00 same 3h-band as window", morning, 0, timeStringToSlot("11:00"), false},
		{"morning on wrong day", morning, 1, timeStringToSlot("09:30"), false},
		{"all day at midnight", allDay, 1, timeStringToSlot("00:00"), true},
		{"all day at 23:55", allDay, 1, timeStringToSlot("23:55"), true},
		{"all day on wrong day", allDay, 0, timeStringToSlot("12:00"), false},
		{"overnight at 23:00", overnight, 0, timeStringToSlot("23:00"), true},
		{"overnight at 05:00", overnight, 0, timeStringToSlot("05:00"), true},
		{"overnight at 12:00", overnight, 0, timeStringToSlot("12:00"), false},
		{"inactive schedule never covers", inactive, 0, timeStringToSlot("12:00"), false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.window.coversSlot(tc.day, tc.slot)
			if got != tc.want {
				t.Errorf("coversSlot(%d, %d) = %v; want %v", tc.day, tc.slot, got, tc.want)
			}
		})
	}
}

func TestScheduleCache_LoadScheduleRules(t *testing.T) {
	sc := NewScheduleCache().(*scheduleCache)
	directRules := []*model.ScheduleRule{
		{ScheduleID: 1, RuleID: 100},
		{ScheduleID: 1, RuleID: 101},
	}
	scheduleLists := []*model.ScheduleList{
		{ScheduleID: 1, ListID: 10},
		{ScheduleID: 2, ListID: 10},
	}
	listRules := []*model.ListRule{
		{ListID: 10, RuleID: 200},
		{ListID: 10, RuleID: 201},
	}

	sc.LoadScheduleRules(directRules, scheduleLists, listRules)

	if len(sc.scheduleRuleSet[1]) != 2 {
		t.Errorf("scheduleRuleSet[1] has %d entries; want 2", len(sc.scheduleRuleSet[1]))
	}
	if len(sc.scheduleRuleSet[2]) != 0 {
		t.Errorf("scheduleRuleSet[2] has %d entries; want 0", len(sc.scheduleRuleSet[2]))
	}
	if len(sc.scheduleToLists[1]) != 1 || sc.scheduleToLists[1][0] != 10 {
		t.Errorf("scheduleToLists[1] = %v; want [10]", sc.scheduleToLists[1])
	}
	if len(sc.scheduleToLists[2]) != 1 || sc.scheduleToLists[2][0] != 10 {
		t.Errorf("scheduleToLists[2] = %v; want [10]", sc.scheduleToLists[2])
	}
	if len(sc.listRuleSet[10]) != 2 {
		t.Errorf("listRuleSet[10] has %d entries; want 2", len(sc.listRuleSet[10]))
	}
}

func TestScheduleCache_HasRuleIntersection(t *testing.T) {
	sc := NewScheduleCache()
	directRules := []*model.ScheduleRule{
		{ScheduleID: 1, RuleID: 10},
		{ScheduleID: 1, RuleID: 20},
		{ScheduleID: 2, RuleID: 30},
	}
	scheduleLists := []*model.ScheduleList{
		{ScheduleID: 3, ListID: 100},
		{ScheduleID: 4, ListID: 100},
		{ScheduleID: 4, ListID: 200},
	}
	listRules := []*model.ListRule{
		{ListID: 100, RuleID: 40},
		{ListID: 100, RuleID: 50},
		{ListID: 200, RuleID: 60},
	}
	sc.LoadScheduleRules(directRules, scheduleLists, listRules)

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
		{"Intersection through shared list", []int{3}, []int{50, 99}, true},
		{"Intersection through second subscribed list", []int{4}, []int{60, 99}, true},
		{"No intersection through shared list", []int{3}, []int{60, 99}, false},
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

	allDays := model.Schedule{
		Monday:    true,
		Tuesday:   true,
		Wednesday: true,
		Thursday:  true,
		Friday:    true,
		Saturday:  true,
		Sunday:    true,
	}

	// An always-active schedule
	s1 := allDays
	s1.ID = 1
	s1.Active = true
	s1.StartTime = "00:00"
	s1.EndTime = "00:00"

	// A never-active schedule (no days)
	s2 := &model.Schedule{
		ID:        2,
		Active:    true,
		StartTime: "00:00",
		EndTime:   "00:00",
	}

	// A deactivated schedule that would otherwise always be active
	s3 := allDays
	s3.ID = 3
	s3.Active = false
	s3.StartTime = "00:00"
	s3.EndTime = "00:00"

	sc.LoadSchedules([]*model.Schedule{&s1, s2, &s3})

	// s1 is guaranteed active (all days, all day), s2 is guaranteed inactive
	// (no days), s3 is guaranteed inactive (active flag unset), 4 is unknown.
	active := sc.FilterActiveSchedules([]int{1, 2, 3, 4})
	if len(active) != 1 || active[0] != 1 {
		t.Errorf("FilterActiveSchedules([1, 2, 3, 4]) = %v; want [1]", active)
	}

	empty := sc.FilterActiveSchedules([]int{})
	if len(empty) != 0 {
		t.Errorf("FilterActiveSchedules([]) = %v; want []", empty)
	}
}

func TestScheduleCache_FilterActiveSchedules_TimeOfDayPrecision(t *testing.T) {
	sc := NewScheduleCache()

	now := time.Now()

	// A one-hour window centered on the current 5-minute slot
	start := now.Add(-30 * time.Minute)
	end := now.Add(30 * time.Minute)

	// Skip the edge case where the window would cross midnight
	if start.Day() != end.Day() {
		t.Skip("current time too close to midnight for this test")
	}

	current := model.Schedule{
		ID:        1,
		Active:    true,
		StartTime: start.Format("15:04"),
		EndTime:   end.Format("15:04"),
		Monday:    true,
		Tuesday:   true,
		Wednesday: true,
		Thursday:  true,
		Friday:    true,
		Saturday:  true,
		Sunday:    true,
	}

	// A window in the same 3-hour band that already ended over an hour ago.
	// With the old 8-bit day masks this was indistinguishable from an active
	// window.
	past := current
	past.ID = 2
	past.StartTime = start.Add(-90 * time.Minute).Format("15:04")
	past.EndTime = end.Add(-90 * time.Minute).Format("15:04")

	if start.Add(-90*time.Minute).Day() != start.Day() {
		t.Skip("current time too close to midnight for this test")
	}

	sc.LoadSchedules([]*model.Schedule{&current, &past})

	active := sc.FilterActiveSchedules([]int{1, 2})
	if len(active) != 1 || active[0] != 1 {
		t.Errorf("FilterActiveSchedules([1, 2]) = %v; want [1]", active)
	}
}
