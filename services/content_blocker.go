package services

import (
	"fmt"
	"regexp"
	"server/repos"
	"strings"
)

// Define ContentBlocker struct
type ContentBlocker struct {
	ScheduleRepo *repos.ScheduleRepoImpl
}

// Define domainRegex to check for valid domains
var domainRegex = regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

// Hardcoded blocked domains (temporarily)
var blockedDomains = []string{
	"reddit.com",
	"startmunich.de",
	"youtube.com",
}

// Check if a given domain is blocked
func (c *ContentBlocker) IsBlocked(domain string) (bool, error) {
	// Check if domain is valid
	if !domainRegex.MatchString(domain) {
		return false, fmt.Errorf("%s is not a valid domain\n", domain)
	}

	// Normalize the domain to lowercase and remove trailing dot if present
	domain = strings.ToLower(strings.TrimSuffix(domain, "."))

	// Check if domain is blocked
	for _, blocked := range blockedDomains {
		if strings.HasSuffix(domain, blocked) {
			return true, nil
		}
	}
	return false, nil
}
