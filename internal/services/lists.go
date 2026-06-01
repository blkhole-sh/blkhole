package services

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/lemon3studio/blkhole/internal/model"
	"github.com/lemon3studio/blkhole/internal/repos"
)

// Regex patterns used across file parsing functions
var (
	validDomainPattern = regexp.MustCompile(`^[a-zA-Z0-9.-]+\.[a-zA-Z0-9-]{2,}$`)
	adblockPattern     = regexp.MustCompile(`^\|\|([a-zA-Z0-9.-]+\.[a-zA-Z0-9-]{2,})\^`)
	whitelistPattern   = regexp.MustCompile(`^@@\|\|([a-zA-Z0-9.-]+\.[a-zA-Z0-9-]{2,})\^`)
	hostsPattern       = regexp.MustCompile(`^(?:0\.0\.0\.0|127\.0\.0\.1)\s+([a-zA-Z0-9.-]+\.[a-zA-Z0-9-]{2,})`)
	httpsPattern       = regexp.MustCompile(`^https?://[^\s]+$`)
	localFilePattern   = regexp.MustCompile(`^(\.?\.?/|/)[^\s]*$`)
	whitespacePattern  = regexp.MustCompile(`\s+`)
	trimPattern        = regexp.MustCompile(`^\s+|\s+$`)
)

// ListService defines the interface for list service operations
type ListService interface {
	LoadList(*model.List) error
}

// listService implements the ListService interface
type listService struct {
	lists          repos.ListRepo
	ruleRepo       repos.RuleRepo
	domainRepo     repos.DomainRepo
	contentBlocker ContentBlocker
}

// NewListService creates a new ListService instance
func NewListService(lists repos.ListRepo, ruleRepo repos.RuleRepo, domainRepo repos.DomainRepo, contentBlocker ContentBlocker) ListService {
	return &listService{
		lists:          lists,
		ruleRepo:       ruleRepo,
		domainRepo:     domainRepo,
		contentBlocker: contentBlocker,
	}
}

func readAdblockFile(r io.Reader, domainRepo repos.DomainRepo) ([]model.Rule, error) {
	var rules []model.Rule
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		line = whitespacePattern.ReplaceAllString(line, " ") // normalize whitespace
		line = trimPattern.ReplaceAllString(line, "")        // trim

		// Skip empty lines, comments, and adblock headers
		if line == "" || (len(line) > 0 && (line[0] == '#' || line[0] == '!' || line[0] == '[')) {
			continue
		}

		var domain string
		var allowed bool

		// Check for whitelist entries first (@@||domain^)
		if matches := whitelistPattern.FindStringSubmatch(line); matches != nil {
			domain = matches[1]
			allowed = true
		} else if matches := adblockPattern.FindStringSubmatch(line); matches != nil {
			// Adblock format (||domain^)
			domain = matches[1]
			allowed = false
		} else {
			continue
		}

		// Validate domain format
		if !validDomainPattern.MatchString(domain) {
			if len(rules) < 5 {
				log.Printf("warning: invalid domain format in adblock file: %s", domain)
			}
			continue
		}

		// Normalize domain by removing www. prefix
		domain = strings.TrimPrefix(domain, "www.")

		// Create or get domain first
		domainModel := &model.Domain{Name: domain}
		domainID, err := domainRepo.CreateOrGet(domainModel)
		if err != nil {
			log.Printf("error creating domain %s: %v", domain, err)
			continue
		}

		// Create rule
		rule := model.Rule{
			DomainID: domainID,
			Allowed:  allowed,
		}
		rules = append(rules, rule)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading adblock file: %w", err)
	}

	return rules, nil
}

func readHostFile(r io.Reader, domainRepo repos.DomainRepo) ([]model.Rule, error) {
	var rules []model.Rule
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		line = whitespacePattern.ReplaceAllString(line, " ") // normalize whitespace
		line = trimPattern.ReplaceAllString(line, "")        // trim

		// Skip empty lines and comments
		if line == "" || (len(line) > 0 && line[0] == '#') {
			continue
		}

		// Check for hosts format (0.0.0.0 domain or 127.0.0.1 domain)
		if matches := hostsPattern.FindStringSubmatch(line); matches != nil {
			domain := matches[1]

			// Skip localhost entries
			if domain == "localhost" || domain == "localhost.localdomain" {
				continue
			}

			// Validate domain format
			if !validDomainPattern.MatchString(domain) {
				log.Printf("warning: invalid domain format in hosts file: %s", domain)
				continue
			}

			// Normalize domain by removing www. prefix
			domain = strings.TrimPrefix(domain, "www.")

			// Create or get domain first
			domainModel := &model.Domain{Name: domain}
			domainID, err := domainRepo.CreateOrGet(domainModel)
			if err != nil {
				log.Printf("error creating domain %s: %v", domain, err)
				continue
			}

			// Create rule (hosts files are always blocking)
			rule := model.Rule{
				DomainID: domainID,
				Allowed:  false,
			}
			rules = append(rules, rule)
		} else {
			// Log warning for unrecognized format
			log.Printf("warning: unrecognized hosts format in line: %s", line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading hosts file: %w", err)
	}

	return rules, nil
}

func readDomainsFile(r io.Reader, domainRepo repos.DomainRepo) ([]model.Rule, error) {
	var rules []model.Rule
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		line = whitespacePattern.ReplaceAllString(line, " ") // normalize whitespace
		line = trimPattern.ReplaceAllString(line, "")        // trim

		// Skip empty lines and comments
		if line == "" || (len(line) > 0 && line[0] == '#') {
			continue
		}

		// Validate domain format (line should be just a domain)
		if !validDomainPattern.MatchString(line) {
			log.Printf("invalid domain format in domains file: %s", line)
			continue
		}

		// Normalize domain by removing www. prefix
		domain := strings.TrimPrefix(line, "www.")

		// Create or get domain first
		domainModel := &model.Domain{Name: domain}
		domainID, err := domainRepo.CreateOrGet(domainModel)
		if err != nil {
			log.Printf("error creating domain %s: %v", line, err)
			continue
		}

		// Create rule (default to blocking for domains file)
		rule := model.Rule{
			DomainID: domainID,
			Allowed:  false,
		}
		rules = append(rules, rule)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading domains file: %w", err)
	}

	return rules, nil
}

// detectAndReadFile auto-detects file format and calls appropriate parser
func detectAndReadFile(r io.Reader, domainRepo repos.DomainRepo) ([]model.Rule, error) {
	// Read entire file content
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("error reading file for format detection: %w", err)
	}

	contentStr := string(content)

	if len(contentStr) == 0 {
		log.Printf("warning: file appears to be empty")
		return []model.Rule{}, nil
	}

	// Split content into lines for more accurate pattern detection
	lines := strings.Split(contentStr, "\n")

	var adblockCount, hostsCount, domainCount int

	// Sample first 100 lines for pattern detection (or all lines if file is smaller)
	sampleSize := min(100, len(lines))

	for i := range sampleSize {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") || strings.HasPrefix(line, "[") {
			continue
		}

		// Check for adblock patterns
		if adblockPattern.MatchString(line) || whitelistPattern.MatchString(line) {
			adblockCount++
		} else if hostsPattern.MatchString(line) {
			hostsCount++
		} else if validDomainPattern.MatchString(line) {
			domainCount++
		}
	}

	// Determine file type based on pattern counts
	if adblockCount > 0 {
		return readAdblockFile(strings.NewReader(contentStr), domainRepo)
	} else if hostsCount > 0 {
		return readHostFile(strings.NewReader(contentStr), domainRepo)
	} else if domainCount > 0 {
		return readDomainsFile(strings.NewReader(contentStr), domainRepo)
	} else {
		log.Printf("warning: unable to detect file format - no recognizable patterns found in first %d lines", sampleSize)
		return []model.Rule{}, nil
	}
}

// readHTTPSFile reads rules from an HTTPS file.
// We use http.DefaultClient to allow test overrides via httptest.Client().
func readHTTPSFile(url string, domainRepo repos.DomainRepo) ([]model.Rule, error) {
	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to GET url %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	return detectAndReadFile(resp.Body, domainRepo)
}

// readLocalFile reads rules from a local file.
func readLocalFile(path string, domainRepo repos.DomainRepo) ([]model.Rule, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open local file %s: %w", path, err)
	}
	defer file.Close()

	return detectAndReadFile(file, domainRepo)
}

func readFromSource(source string, domainRepo repos.DomainRepo) ([]model.Rule, error) {
	switch {
	case httpsPattern.MatchString(source):
		return readHTTPSFile(source, domainRepo)
	case localFilePattern.MatchString(source):
		return readLocalFile(source, domainRepo)
	default:
		return nil, fmt.Errorf("neither HTTPS file, nor local file")
	}
}

func (ls *listService) LoadList(l *model.List) error {
	// Check if no source to load from
	if l.Source == "" {
		return nil
	}

	// Read rules from the source
	rules, err := readFromSource(l.Source, ls.domainRepo)
	if err != nil {
		return fmt.Errorf("failed to read from source %s: %w", l.Source, err)
	}

	// Prepare rules for batch creation
	rulePtrs := make([]*model.Rule, len(rules))
	for i := range rules {
		rulePtrs[i] = &rules[i]
	}

	// Batch create or get rules
	if err := ls.ruleRepo.BatchCreateOrGet(rulePtrs); err != nil {
		return fmt.Errorf("failed to batch create rules: %w", err)
	}

	// Extract rule IDs for linking
	ruleIDs := make([]int, len(rules))
	for i, r := range rules {
		ruleIDs[i] = r.ID
	}

	// Batch link rules to list
	if err := ls.ruleRepo.BatchLinkToList(ruleIDs, l.ID); err != nil {
		return fmt.Errorf("failed to batch link rules to list %d: %w", l.ID, err)
	}

	// Load the updated rules into the list model
	l.Rules, err = ls.lists.LoadRules(l.ID)
	if err != nil {
		return fmt.Errorf("failed to load rules for list %d: %w", l.ID, err)
	}

	l.Count = len(l.Rules)

	// Reload the content blocker cache to pick up the new domains
	if err := ls.contentBlocker.Reload(); err != nil {
		return fmt.Errorf("failed to reload content blocker after loading list %d: %w", l.ID, err)
	}

	return nil
}
