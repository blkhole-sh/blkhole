// Package cache provides in-memory caching for DNS blocking lookups
package cache

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lemon3studio/blkhole/internal/model"
)

const (
	slotsPerHour = 12 // 5-minute slots per hour (60/5)
	bitsPerDay   = 8  // Each day gets 8 bits in the 64-bit mask (64 bits / 8 days = 8 bits per day)
)

// ScheduleCache provides fast schedule-to-rule lookups and rule intersection checks
type ScheduleCache interface {
	LoadSchedules(schedules []*model.Schedule)
	LoadScheduleRules(scheduleRules []*model.ScheduleRule)
	GetRules(scheduleID int) []int
	HasRuleIntersection(scheduleIDs []int, domainRules []int) bool
	FilterActiveSchedules(scheduleIDs []int) []int
}

// scheduleCache implements the ScheduleCache interface
type scheduleCache struct {
	mu              sync.RWMutex
	scheduleToRule  map[int][]int            // Schedule ID → Rule IDs
	scheduleRuleSet map[int]map[int]struct{} // Schedule ID → Rule IDs as hash set for O(1) lookup
	scheduleMasks   map[int]uint64           // Schedule ID → Pre-computed bitmask for time filtering
}

// NewScheduleCache creates a new schedule cache instance
func NewScheduleCache() ScheduleCache {
	return &scheduleCache{
		scheduleToRule:  make(map[int][]int),
		scheduleRuleSet: make(map[int]map[int]struct{}),
		scheduleMasks:   make(map[int]uint64),
	}
}

// LoadSchedules populates the schedule cache with pre-computed bitmasks for active filtering
func (sc *scheduleCache) LoadSchedules(schedules []*model.Schedule) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	for _, schedule := range schedules {
		sc.scheduleMasks[schedule.ID] = convertToBitmask(schedule)
	}
}

// convertToBitmask converts Schedule boolean days and time strings to efficient bitmask
func convertToBitmask(schedule *model.Schedule) uint64 {
	var mask uint64

	// Convert day booleans to bitmask (Monday=bit0, Tuesday=bit1, ..., Sunday=bit6)
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
			days |= 1 << i // Set bit position i for active day
		}
	}

	// Convert time strings to slots (0-287, where each slot = 5 minutes)
	startSlot := timeStringToSlot(schedule.StartTime)
	endSlot := timeStringToSlot(schedule.EndTime)

	// Create bitmask for each active day (each day gets 8 bits: Monday=0-7, Tuesday=8-15, etc.)
	for day := range 7 {
		if days&(1<<day) != 0 {
			mask |= calculateDayMask(uint16(day), startSlot, endSlot)
		}
	}

	return mask
}

// calculateDayMask generates the bitmask for a specific day given start and end slots
func calculateDayMask(day uint16, startSlot, endSlot uint16) uint64 {
	var dayMask uint64
	dayOffset := day * bitsPerDay

	for slot := startSlot; slot <= endSlot; slot++ {
		bitInDay := (slot * bitsPerDay) / (24 * slotsPerHour)
		if bitInDay < bitsPerDay {
			dayMask |= 1 << (dayOffset + bitInDay)
		}
	}

	return dayMask
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

// getCurrentTimeMask creates a bitmask for the current time
func getCurrentTimeMask() uint64 {
	now := time.Now()

	// Convert Go's Sunday=0 weekday to Monday=0 format
	dayBit := uint8((now.Weekday() + 6) % 7)

	// Convert current time to slot number (0-287)
	currentSlot := uint16(now.Hour()*slotsPerHour + now.Minute()/5)

	// Calculate bit position for current day and time
	dayOffset := uint16(dayBit) * bitsPerDay

	// Scale current slot to bit position within the day (0-7)
	bitInDay := (currentSlot * bitsPerDay) / (24 * slotsPerHour)
	if bitInDay < bitsPerDay {
		return 1 << (dayOffset + bitInDay)
	}

	return 0
}

// LoadScheduleRules populates the schedule-to-rule mapping
func (sc *scheduleCache) LoadScheduleRules(scheduleRules []*model.ScheduleRule) {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	for _, sr := range scheduleRules {
		// Add to slice (keep for compatibility)
		sc.scheduleToRule[sr.ScheduleID] = append(sc.scheduleToRule[sr.ScheduleID], sr.RuleID)

		// Add to hash set for O(1) lookup
		if sc.scheduleRuleSet[sr.ScheduleID] == nil {
			sc.scheduleRuleSet[sr.ScheduleID] = make(map[int]struct{})
		}
		sc.scheduleRuleSet[sr.ScheduleID][sr.RuleID] = struct{}{}
	}
}

// GetRules returns all rule IDs for a given schedule ID
func (sc *scheduleCache) GetRules(scheduleID int) []int {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	return sc.scheduleToRule[scheduleID]
}

// HasRuleIntersection efficiently checks if any schedule rule intersects with domain rules
func (sc *scheduleCache) HasRuleIntersection(scheduleIDs []int, domainRules []int) bool {
	if len(scheduleIDs) == 0 || len(domainRules) == 0 {
		return false
	}

	sc.mu.RLock()
	defer sc.mu.RUnlock()

	// Use schedule rule sets for O(1) lookup
	for _, schedID := range scheduleIDs {
		if ruleSet := sc.scheduleRuleSet[schedID]; ruleSet != nil {
			// Check each domain rule against schedule rule set
			for _, ruleID := range domainRules {
				if _, exists := ruleSet[ruleID]; exists {
					return true
				}
			}
		}
	}
	return false
}

// FilterActiveSchedules returns only the schedules that are currently active using efficient bitmask comparison
func (sc *scheduleCache) FilterActiveSchedules(scheduleIDs []int) []int {
	if len(scheduleIDs) == 0 {
		return nil
	}

	sc.mu.RLock()
	defer sc.mu.RUnlock()

	// Pre-allocate slice with capacity to avoid reallocations
	activeScheduleIDs := make([]int, 0, len(scheduleIDs))

	// Get current time as bitmask for comparison
	currentTimeMask := getCurrentTimeMask()

	// Check each schedule's bitmask against current time
	for _, scheduleID := range scheduleIDs {
		if scheduleMask, exists := sc.scheduleMasks[scheduleID]; exists && scheduleMask&currentTimeMask != 0 {
			activeScheduleIDs = append(activeScheduleIDs, scheduleID)
		}
	}

	return activeScheduleIDs
}
