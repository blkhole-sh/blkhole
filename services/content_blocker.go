package services

import (
	"fmt"
	"log"
	"regexp"
	"server/repos"
	"strings"
)

// Define ContentBlocker struct
type ContentBlocker struct {
	ScheduleRepo *repos.ScheduleRepoImpl
}

// Create new ContentBlocker
func NewContentBlocker(scheduleRepo *repos.ScheduleRepoImpl) *ContentBlocker {
	return &ContentBlocker{ScheduleRepo: scheduleRepo}
}

// Define domainRegex to check for valid domains
var domainRegex = regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

// Blocked DoH providers (in order to force Chromium to use Leo DoH)
var dohBlocklist = []string{
	"dns.google",
	"dns.google.com",
	"cloudflare-dns.com",
	"one.one.one.one",
	"dns.quad9.net",
	"doh.opendns.com",
	"dns.opendns.com",
	"doh.cleanbrowsing.org",
	"dns.nextdns.io",
	"dns.adguard.com",
	"doh.neustar.biz",
	"doh.xfinity.com",
	"doh.cira.ca",
	"doh.yandex.net",
	"doh.powerdns.org",
	"dns.alidns.com",
	"doh.dnssec.works",
}

// Hardcoded blocked domains (temporarily)
var blockedDomains = []string{
	"reddit.com",
	"startmunich.de",
	"youtube.com",
	"dns.google",
}

// Check if a given domain is blocked
func (c *ContentBlocker) IsBlocked(domain string) (bool, error) {
	log.Printf("ContentBlocker | IsBlocked | %s", domain)

	// Check if domain is valid
	if !domainRegex.MatchString(domain) {
		return true, fmt.Errorf("%s is not a valid domain\n", domain)
	}

	// Normalize the domain to lowercase and remove trailing dot if present
	domain = strings.ToLower(strings.TrimSuffix(domain, "."))

	// Strip off .localdomain suffix if present
	domain = strings.TrimSuffix(domain, ".localdomain")

	// Check if domain is blocked
	for _, blocked := range append(dohBlocklist, blockedDomains...) {
		if domain == blocked || strings.HasSuffix(domain, "."+blocked) {
			return true, nil
		}
	}
	return false, nil
}
