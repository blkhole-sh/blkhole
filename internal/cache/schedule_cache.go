// Package cache provides in-memory caching for DNS blocking lookups
package cache

import (
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/blkhole-sh/blkhole/internal/model"
)

const slotsPerHour = 12 // 5-minute slots per hour (60/5)

// ScheduleCache provides fast schedule-to-rule lookups and rule intersection checks
type ScheduleCache interface {
	Load(schedules []*model.Schedule, scheduleRules []*model.ScheduleRule, scheduleLists []*model.ScheduleList, listRules []*model.ListRule)
	LoadSchedules(schedules []*model.Schedule)
	LoadScheduleRules(scheduleRules []*model.ScheduleRule, scheduleLists []*model.ScheduleList, listRules []*model.ListRule)
	HasRuleIntersection(scheduleIDs []int, domainRules []int) bool
	EffectiveRuleIDs(scheduleIDs []int) []int
	FilterActiveSchedules(scheduleIDs []int) []int
}

// EffectiveRuleIDs returns the union of direct and list rules attached to schedules.
func (sc *scheduleCache) EffectiveRuleIDs(scheduleIDs []int) []int {
	snapshot := sc.current()
	ruleIDs := make(map[int]struct{})
	for _, scheduleID := range scheduleIDs {
		for ruleID := range snapshot.scheduleRuleSet[scheduleID] {
			ruleIDs[ruleID] = struct{}{}
		}
		for _, listID := range snapshot.scheduleToLists[scheduleID] {
			for ruleID := range snapshot.listRuleSet[listID] {
				ruleIDs[ruleID] = struct{}{}
			}
		}
	}

	result := make([]int, 0, len(ruleIDs))
	for ruleID := range ruleIDs {
		result = append(result, ruleID)
	}
	return result
}

// scheduleWindow holds a schedule's active days and time-of-day window with
// 5-minute slot precision
type scheduleWindow struct {
	active    bool
	days      uint8 // Monday=bit0, Tuesday=bit1, ..., Sunday=bit6
	allDay    bool
	startSlot uint16 // inclusive, 0-287
	endSlot   uint16 // inclusive, 0-287
}

type scheduleSnapshot struct {
	scheduleRuleSet map[int]map[int]struct{} // Schedule ID → directly attached Rule IDs
	scheduleToLists map[int][]int            // Schedule ID → subscribed List IDs
	listRuleSet     map[int]map[int]struct{} // List ID → shared Rule IDs
	scheduleWindows map[int]scheduleWindow   // Schedule ID → Pre-computed window for time filtering
}

// scheduleCache implements the ScheduleCache interface
type scheduleCache struct {
	writeMu  sync.Mutex
	snapshot atomic.Pointer[scheduleSnapshot]
}

// NewScheduleCache creates a new schedule cache instance
func NewScheduleCache() ScheduleCache {
	sc := &scheduleCache{}
	sc.snapshot.Store(newScheduleSnapshot())
	return sc
}

func newScheduleSnapshot() *scheduleSnapshot {
	return &scheduleSnapshot{
		scheduleRuleSet: make(map[int]map[int]struct{}),
		scheduleToLists: make(map[int][]int),
		listRuleSet:     make(map[int]map[int]struct{}),
		scheduleWindows: make(map[int]scheduleWindow),
	}
}

func buildScheduleWindows(schedules []*model.Schedule) map[int]scheduleWindow {
	scheduleWindows := make(map[int]scheduleWindow)
	for _, schedule := range schedules {
		scheduleWindows[schedule.ID] = convertToWindow(schedule)
	}
	return scheduleWindows
}

func buildScheduleRuleSet(scheduleRules []*model.ScheduleRule) map[int]map[int]struct{} {
	scheduleRuleSet := make(map[int]map[int]struct{})
	for _, sr := range scheduleRules {
		if scheduleRuleSet[sr.ScheduleID] == nil {
			scheduleRuleSet[sr.ScheduleID] = make(map[int]struct{})
		}
		scheduleRuleSet[sr.ScheduleID][sr.RuleID] = struct{}{}
	}
	return scheduleRuleSet
}

func buildScheduleToLists(scheduleLists []*model.ScheduleList) map[int][]int {
	scheduleToLists := make(map[int][]int)
	for _, sl := range scheduleLists {
		scheduleToLists[sl.ScheduleID] = append(scheduleToLists[sl.ScheduleID], sl.ListID)
	}
	return scheduleToLists
}

func buildListRuleSet(listRules []*model.ListRule) map[int]map[int]struct{} {
	listRuleSet := make(map[int]map[int]struct{})
	for _, lr := range listRules {
		if listRuleSet[lr.ListID] == nil {
			listRuleSet[lr.ListID] = make(map[int]struct{})
		}
		listRuleSet[lr.ListID][lr.RuleID] = struct{}{}
	}
	return listRuleSet
}

func (sc *scheduleCache) current() *scheduleSnapshot {
	if snapshot := sc.snapshot.Load(); snapshot != nil {
		return snapshot
	}
	return newScheduleSnapshot()
}

// Load publishes a complete schedule cache snapshot.
func (sc *scheduleCache) Load(schedules []*model.Schedule, scheduleRules []*model.ScheduleRule, scheduleLists []*model.ScheduleList, listRules []*model.ListRule) {
	sc.writeMu.Lock()
	defer sc.writeMu.Unlock()

	sc.snapshot.Store(&scheduleSnapshot{
		scheduleRuleSet: buildScheduleRuleSet(scheduleRules),
		scheduleToLists: buildScheduleToLists(scheduleLists),
		listRuleSet:     buildListRuleSet(listRules),
		scheduleWindows: buildScheduleWindows(schedules),
	})
}

// LoadSchedules populates the schedule cache with pre-computed windows for active filtering
func (sc *scheduleCache) LoadSchedules(schedules []*model.Schedule) {
	sc.writeMu.Lock()
	defer sc.writeMu.Unlock()

	current := sc.current()
	sc.snapshot.Store(&scheduleSnapshot{
		scheduleRuleSet: current.scheduleRuleSet,
		scheduleToLists: current.scheduleToLists,
		listRuleSet:     current.listRuleSet,
		scheduleWindows: buildScheduleWindows(schedules),
	})
}

// convertToWindow converts Schedule boolean days and time strings to a scheduleWindow
func convertToWindow(schedule *model.Schedule) scheduleWindow {
	activeDays := []bool{
		schedule.Monday,
		schedule.Tuesday,
		schedule.Wednesday,
		schedule.Thursday,
		schedule.Friday,
		schedule.Saturday,
		schedule.Sunday,
	}

	var days uint8
	for i, active := range activeDays {
		if active {
			days |= 1 << i
		}
	}

	return scheduleWindow{
		active: schedule.Active,
		days:   days,
		// Equal start and end time means all day
		allDay:    schedule.StartTime == schedule.EndTime,
		startSlot: timeStringToSlot(schedule.StartTime),
		endSlot:   timeStringToSlot(schedule.EndTime),
	}
}

// coversSlot reports whether the window covers the given day (Monday=0) and
// 5-minute slot. Windows whose end slot is before their start slot wrap past
// midnight (e.g. 22:00-06:00).
func (w scheduleWindow) coversSlot(day uint8, slot uint16) bool {
	if !w.active || w.days&(1<<day) == 0 {
		return false
	}
	if w.allDay {
		return true
	}
	if w.startSlot <= w.endSlot {
		return w.startSlot <= slot && slot <= w.endSlot
	}
	// Overnight window
	return slot >= w.startSlot || slot <= w.endSlot
}

// timeStringToSlot converts "HH:MM" format to 5-minute slot number (0-287)
func timeStringToSlot(timeStr string) uint16 {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0
	}

	// Parse hour component (0-23)
	hour, err := strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour >= 24 {
		return 0
	}

	// Parse minute component (0-59)
	minute, err := strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute >= 60 {
		return 0
	}

	// Convert to slot: hour*12 + minute/5 (e.g., "09:15" -> 9*12 + 15/5 = 111)
	return uint16(hour*slotsPerHour + minute/5)
}

// LoadScheduleRules populates direct schedule rules and shared list rule mappings
func (sc *scheduleCache) LoadScheduleRules(scheduleRules []*model.ScheduleRule, scheduleLists []*model.ScheduleList, listRules []*model.ListRule) {
	sc.writeMu.Lock()
	defer sc.writeMu.Unlock()

	current := sc.current()
	sc.snapshot.Store(&scheduleSnapshot{
		scheduleRuleSet: buildScheduleRuleSet(scheduleRules),
		scheduleToLists: buildScheduleToLists(scheduleLists),
		listRuleSet:     buildListRuleSet(listRules),
		scheduleWindows: current.scheduleWindows,
	})
}

// hasMatchingRule checks if any rule in domainRules exists in the provided ruleSet
func hasMatchingRule(ruleSet map[int]struct{}, domainRules []int) bool {
	for _, ruleID := range domainRules {
		if _, exists := ruleSet[ruleID]; exists {
			return true
		}
	}
	return false
}

// HasRuleIntersection efficiently checks if any schedule rule intersects with domain rules
func (sc *scheduleCache) HasRuleIntersection(scheduleIDs []int, domainRules []int) bool {
	if len(scheduleIDs) == 0 || len(domainRules) == 0 {
		return false
	}

	snapshot := sc.current()

	for _, schedID := range scheduleIDs {
		if ruleSet := snapshot.scheduleRuleSet[schedID]; ruleSet != nil {
			if hasMatchingRule(ruleSet, domainRules) {
				return true
			}
		}

		for _, listID := range snapshot.scheduleToLists[schedID] {
			if ruleSet := snapshot.listRuleSet[listID]; ruleSet != nil {
				if hasMatchingRule(ruleSet, domainRules) {
					return true
				}
			}
		}
	}
	return false
}

// FilterActiveSchedules returns only the schedules that are active at the current time
func (sc *scheduleCache) FilterActiveSchedules(scheduleIDs []int) []int {
	if len(scheduleIDs) == 0 {
		return nil
	}

	snapshot := sc.current()

	now := time.Now()

	// Convert Go's Sunday=0 weekday to Monday=0 format
	day := uint8((now.Weekday() + 6) % 7)

	// Convert current time to slot number (0-287)
	slot := uint16(now.Hour()*slotsPerHour + now.Minute()/5)

	// Pre-allocate slice with capacity to avoid reallocations
	activeScheduleIDs := make([]int, 0, len(scheduleIDs))

	for _, scheduleID := range scheduleIDs {
		if window, exists := snapshot.scheduleWindows[scheduleID]; exists && window.coversSlot(day, slot) {
			activeScheduleIDs = append(activeScheduleIDs, scheduleID)
		}
	}

	return activeScheduleIDs
}
