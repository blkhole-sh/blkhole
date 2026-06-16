// Package services provides business logic services for the blkhole DNS blocker application.
package services

import (
	"fmt"
	"regexp"
	"strings"
	"sync/atomic"

	"github.com/blkhole-sh/blkhole/internal/cache"
	"github.com/blkhole-sh/blkhole/internal/repos"
)

// domainRegex is used to check for valid domain format
var domainRegex = regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

// ContentBlocker defines the interface for content blocking operations
type ContentBlocker interface {
	Init() error
	Reload() error
	IsBlocked(domain string, deviceHash string) (bool, error)
}

type contentBlockerSnapshot struct {
	domainCache   cache.DomainCache
	deviceCache   cache.DeviceCache
	scheduleCache cache.ScheduleCache
}

// contentBlocker implements the ContentBlocker interface
type contentBlocker struct {
	devices     repos.DeviceRepo
	rules       repos.RuleRepo
	schedules   repos.ScheduleRepo
	domains     repos.DomainRepo
	deviceCache cache.DeviceCache
	caches      atomic.Pointer[contentBlockerSnapshot]
}

// NewContentBlocker creates a new ContentBlocker instance
func NewContentBlocker(devices repos.DeviceRepo, rules repos.RuleRepo, schedules repos.ScheduleRepo, domains repos.DomainRepo, deviceCache cache.DeviceCache) ContentBlocker {
	cb := &contentBlocker{
		devices:     devices,
		rules:       rules,
		schedules:   schedules,
		domains:     domains,
		deviceCache: deviceCache,
	}
	cb.caches.Store(&contentBlockerSnapshot{
		domainCache:   cache.NewDomainCache(),
		deviceCache:   deviceCache,
		scheduleCache: cache.NewScheduleCache(),
	})
	return cb
}

func (cb *contentBlocker) currentCaches() *contentBlockerSnapshot {
	if caches := cb.caches.Load(); caches != nil {
		return caches
	}
	return &contentBlockerSnapshot{
		domainCache:   cache.NewDomainCache(),
		deviceCache:   cache.NewDeviceCache(),
		scheduleCache: cache.NewScheduleCache(),
	}
}

func (cb *contentBlocker) buildCacheSnapshot() (*contentBlockerSnapshot, error) {
	domains, err := cb.domains.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to load domains: %w", err)
	}
	rules, err := cb.rules.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to load rules: %w", err)
	}
	domainCache := cache.NewDomainCache()
	domainCache.Load(domains, rules)

	schedules, err := cb.schedules.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to load schedules: %w", err)
	}
	scheduleRules, err := cb.schedules.FindScheduleRule()
	if err != nil {
		return nil, fmt.Errorf("failed to load schedule rules: %w", err)
	}
	scheduleLists, err := cb.schedules.FindScheduleList()
	if err != nil {
		return nil, fmt.Errorf("failed to load schedule lists: %w", err)
	}
	listRules, err := cb.schedules.FindListRule()
	if err != nil {
		return nil, fmt.Errorf("failed to load list rules: %w", err)
	}
	scheduleCache := cache.NewScheduleCache()
	scheduleCache.Load(schedules, scheduleRules, scheduleLists, listRules)

	devices, err := cb.devices.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to load devices: %w", err)
	}
	deviceSchedules, err := cb.devices.FindDeviceSchedule()
	if err != nil {
		return nil, fmt.Errorf("failed to load device schedules: %w", err)
	}
	deviceCache := cache.NewDeviceCache()
	deviceCache.Load(devices, deviceSchedules)
	cb.deviceCache.Load(devices, deviceSchedules)

	return &contentBlockerSnapshot{
		domainCache:   domainCache,
		deviceCache:   deviceCache,
		scheduleCache: scheduleCache,
	}, nil
}

// Reload clears and rebuilds the content blocker cache from the database
func (cb *contentBlocker) Reload() error {
	return cb.Init()
}

// Init initializes the content blocker by loading data from database into cache modules
func (cb *contentBlocker) Init() error {
	caches, err := cb.buildCacheSnapshot()
	if err != nil {
		return err
	}
	cb.caches.Store(caches)
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
	caches := cb.currentCaches()

	deviceID, ok := caches.deviceCache.GetDeviceID(deviceHash)
	if !ok {
		// Unknown devices are not blocked
		return false, nil
	}

	// 2. Get schedules for device and filter for currently active ones
	scheduleIDs := caches.deviceCache.GetSchedules(deviceID)
	if len(scheduleIDs) == 0 {
		return false, nil
	}

	// Filter to only schedules that are active right now
	activeScheduleIDs := caches.scheduleCache.FilterActiveSchedules(scheduleIDs)
	if len(activeScheduleIDs) == 0 {
		return false, nil
	}

	// 3. Resolve domain to ID using longest-prefix match
	domainID, ok := caches.domainCache.LookupDomainID(domain)
	if !ok {
		return false, nil
	}

	// 4. Get rules for the domain
	domainRules := caches.domainCache.GetRules(domainID)
	if len(domainRules) == 0 {
		return false, nil
	}

	// 5. Check for intersection: active schedules → rules → domain rules
	result := caches.scheduleCache.HasRuleIntersection(activeScheduleIDs, domainRules)
	return result, nil
}
