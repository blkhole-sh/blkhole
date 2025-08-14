// Package services provides business logic services for the Leo DNS blocker application.
package services

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/lemon3studio/leo/internal/cache"
	"github.com/lemon3studio/leo/internal/repos"
)

// domainRegex is used to check for valid domain format
var domainRegex = regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

// ContentBlocker defines the interface for content blocking operations
type ContentBlocker interface {
	Init() error
	IsBlocked(domain string, deviceHash string) (bool, error)
}

// contentBlocker implements the ContentBlocker interface
type contentBlocker struct {
	devices       repos.DeviceRepo
	rules         repos.RuleRepo
	schedules     repos.ScheduleRepo
	domains       repos.DomainRepo
	domainCache   cache.DomainCache
	deviceCache   cache.DeviceCache
	scheduleCache cache.ScheduleCache
}

// NewContentBlocker creates a new ContentBlocker instance
func NewContentBlocker(devices repos.DeviceRepo, rules repos.RuleRepo, schedules repos.ScheduleRepo, domains repos.DomainRepo) ContentBlocker {
	return &contentBlocker{
		devices:       devices,
		rules:         rules,
		schedules:     schedules,
		domains:       domains,
		domainCache:   cache.NewDomainCache(),
		deviceCache:   cache.NewDeviceCache(),
		scheduleCache: cache.NewScheduleCache(),
	}
}

// initDomains loads all domains into the domain cache
func (cb *contentBlocker) initDomains() error {
	domains, err := cb.domains.FindAll()
	if err != nil {
		return fmt.Errorf("failed to load domains: %w", err)
	}
	cb.domainCache.LoadDomains(domains)
	return nil
}

// initRules loads all rules into the domain cache
func (cb *contentBlocker) initRules() error {
	rules, err := cb.rules.FindAll()
	if err != nil {
		return fmt.Errorf("failed to load rules: %w", err)
	}
	fmt.Printf("DEBUG: Loaded %d rules total\n", len(rules))
	cb.domainCache.LoadRules(rules)
	return nil
}

// initSchedules loads all schedules into the schedule cache
func (cb *contentBlocker) initSchedules() error {
	schedules, err := cb.schedules.FindAll()
	if err != nil {
		return fmt.Errorf("failed to load schedules: %w", err)
	}
	cb.scheduleCache.LoadSchedules(schedules)
	return nil
}

// initScheduleRules loads schedule-rule relationships into the schedule cache
func (cb *contentBlocker) initScheduleRules() error {
	scheduleRules, err := cb.schedules.FindScheduleRule()
	if err != nil {
		return fmt.Errorf("failed to load schedule rules: %w", err)
	}
	cb.scheduleCache.LoadScheduleRules(scheduleRules)
	return nil
}

// initDevices loads all devices into the device cache
func (cb *contentBlocker) initDevices() error {
	devices, err := cb.devices.FindAll()
	if err != nil {
		return fmt.Errorf("failed to load devices: %w", err)
	}
	cb.deviceCache.LoadDevices(devices)
	return nil
}

// initDeviceSchedules loads device-schedule relationships into the device cache
func (cb *contentBlocker) initDeviceSchedules() error {
	deviceSchedules, err := cb.devices.FindDeviceSchedule()
	if err != nil {
		return fmt.Errorf("failed to load device schedules: %w", err)
	}
	cb.deviceCache.LoadDeviceSchedules(deviceSchedules)
	return nil
}

// Init initializes the content blocker by loading data from database into cache modules
func (cb *contentBlocker) Init() error {
	// Load all domains into radix tree for hierarchical domain matching
	if err := cb.initDomains(); err != nil {
		return err
	}

	// Load rules and map them to domains for blocking/allowing decisions
	if err := cb.initRules(); err != nil {
		return err
	}

	// Load schedules and pre-compute bitmasks for efficient time filtering
	if err := cb.initSchedules(); err != nil {
		return err
	}

	// Load schedule-rule relationships for rule intersection checks
	if err := cb.initScheduleRules(); err != nil {
		return err
	}

	// Load devices for device hash to ID mapping
	if err := cb.initDevices(); err != nil {
		return err
	}

	// Load device-schedule relationships to know which schedules apply to which devices
	if err := cb.initDeviceSchedules(); err != nil {
		return err
	}

	return nil
}

// IsBlocked checks if a given domain is blocked
func (cb *contentBlocker) IsBlocked(domain, deviceHash string) (bool, error) {
	// Normalize domain input
	domain = strings.ToLower(strings.TrimSuffix(domain, "."))
	domain = strings.TrimSuffix(domain, ".localdomain")

	// Validate domain format
	if domain == "" || !domainRegex.MatchString(domain) {
		return false, fmt.Errorf("invalid domain format: %s", domain)
	}

	// 1. Resolve device ID
	deviceID, ok := cb.deviceCache.GetDeviceID(deviceHash)
	if !ok {
		// Unknown devices are not blocked
		fmt.Printf("DEBUG: Device hash %s not found\n", deviceHash)
		return false, nil
	}
	fmt.Printf("DEBUG: Device hash %s -> ID %d\n", deviceHash, deviceID)

	// 2. Get schedules for device and filter for currently active ones
	scheduleIDs := cb.deviceCache.GetSchedules(deviceID)
	if len(scheduleIDs) == 0 {
		fmt.Printf("DEBUG: No schedules for device %d\n", deviceID)
		return false, nil
	}
	fmt.Printf("DEBUG: Device %d has schedules: %v\n", deviceID, scheduleIDs)

	// Filter to only active schedules using efficient bitmask comparison
	activeScheduleIDs := cb.scheduleCache.FilterActiveSchedules(scheduleIDs)
	if len(activeScheduleIDs) == 0 {
		fmt.Printf("DEBUG: No active schedules from %v\n", scheduleIDs)
		return false, nil
	}
	fmt.Printf("DEBUG: Active schedules: %v\n", activeScheduleIDs)

	// 3. Resolve domain to ID using longest-prefix match
	domainID, ok := cb.domainCache.LookupDomainID(domain)
	if !ok {
		fmt.Printf("DEBUG: Domain %s not found in cache\n", domain)
		return false, nil
	}
	fmt.Printf("DEBUG: Domain %s -> ID %d\n", domain, domainID)

	// 4. Get rules for the domain
	domainRules := cb.domainCache.GetRules(domainID)
	if len(domainRules) == 0 {
		fmt.Printf("DEBUG: No rules for domain ID %d\n", domainID)
		return false, nil
	}
	fmt.Printf("DEBUG: Domain %d has rules: %v\n", domainID, domainRules)

	// 5. Check for intersection: active schedules → rules → domain rules
	result := cb.scheduleCache.HasRuleIntersection(activeScheduleIDs, domainRules)
	fmt.Printf("DEBUG: Rule intersection check: %v\n", result)
	return result, nil
}
