// Package services provides business logic services for the Leo DNS blocker application.
package services

import (
	"fmt"
	"log"
	"regexp"
	"github.com/lemon3studio/leo/internal/repos"
	"strings"
)

// ContentBlocker defines the interface for content blocking operations
type ContentBlocker interface {
	IsBlocked(domain string, deviceHash string) (bool, error)
}

// contentBlocker implements the ContentBlocker interface
type contentBlocker struct {
	scheduleRepo repos.ScheduleRepo
}

// NewContentBlocker creates a new ContentBlocker instance
func NewContentBlocker(scheduleRepo repos.ScheduleRepo) ContentBlocker {
	return &contentBlocker{scheduleRepo: scheduleRepo}
}

// domainRegex is used to check for valid domain format
var domainRegex = regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

// IsBlocked checks if a given domain is blocked
func (cb *contentBlocker) IsBlocked(domain string, deviceHash string) (bool, error) {
	log.Printf("ContentBlocker | IsBlocked | %s", domain)

	// Check if domain is valid
	if !domainRegex.MatchString(domain) {
		return false, fmt.Errorf("%s is not a valid domain", domain)
	}

	// Normalize the domain to lowercase and remove trailing dot if present
	domain = strings.ToLower(strings.TrimSuffix(domain, "."))

	// Strip off .localdomain suffix if present
	domain = strings.TrimSuffix(domain, ".localdomain")

	// Check if domain is blocked
	return cb.scheduleRepo.DomainBlocked(domain, deviceHash)
}
