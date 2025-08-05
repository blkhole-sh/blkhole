package services

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"github.com/lemon3studio/leo/internal/model"
	"github.com/lemon3studio/leo/internal/repos"
	"strings"
)

// Regex patterns used across file parsing functions
var (
	validDomainPattern = regexp.MustCompile(`^[a-zA-Z0-9.-]+\.[a-zA-Z0-9-]{2,}$`)
	adblockPattern     = regexp.MustCompile(`^\|\|([a-zA-Z0-9.-]+\.[a-zA-Z0-9-]{2,})\^`)
	whitelistPattern   = regexp.MustCompile(`^@@\|\|([a-zA-Z0-9.-]+\.[a-zA-Z0-9-]{2,})\^`)
	hostsPattern       = regexp.MustCompile(`^(?:0\.0\.0\.0|127\.0\.0\.1)\s+([a-zA-Z0-9.-]+\.[a-zA-Z0-9-]{2,})`)
	httpsPattern       = regexp.MustCompile(`^https://[^\s]+$`)
	localFilePattern   = regexp.MustCompile(`^(\.?\.?/|/)[^\s]*$`)
	whitespacePattern  = regexp.MustCompile(`\s+`)
	trimPattern        = regexp.MustCompile(`^\s+|\s+$`)
)

// ListsService defines the interface for list service operations
type ListsService interface {
	LoadList(*model.List) error
}

// listsService implements the ListsService interface
type listsService struct {
	listRepo repos.ListRepo
	ruleRepo repos.RuleRepo
}

// NewListsService creates a new ListsService instance
func NewListsService(listRepo repos.ListRepo, ruleRepo repos.RuleRepo) ListsService {
	return &listsService{
		listRepo: listRepo,
		ruleRepo: ruleRepo,
	}
}

func readAdblockFile(r io.Reader) ([]model.Rule, error) {
	var rules []model.Rule
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		line = whitespacePattern.ReplaceAllString(line, " ") // normalize whitespace
		line = trimPattern.ReplaceAllString(line, "")        // trim

		// Skip empty lines, comments, and adblock headers
		if line == "" || line[0] == '#' || line[0] == '!' || line[0] == '[' {
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
			// Log warning for unrecognized format
			log.Printf("Warning: Unrecognized adblock format in line: %s", line)
			continue
		}

		// Validate domain format
		if !validDomainPattern.MatchString(domain) {
			log.Printf("Warning: Invalid domain format in adblock file: %s", domain)
			continue
		}

		// Create rule
		rule := model.Rule{
			Domain:  domain,
			Allowed: allowed,
			// ListID will be set by the caller
		}
		rules = append(rules, rule)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading adblock file: %w", err)
	}

	return rules, nil
}

func readHostFile(r io.Reader) ([]model.Rule, error) {
	var rules []model.Rule
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		line = whitespacePattern.ReplaceAllString(line, " ") // normalize whitespace
		line = trimPattern.ReplaceAllString(line, "")        // trim

		// Skip empty lines and comments
		if line == "" || line[0] == '#' {
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
				log.Printf("Warning: Invalid domain format in hosts file: %s", domain)
				continue
			}

			// Create rule (hosts files are always blocking)
			rule := model.Rule{
				Domain:  domain,
				Allowed: false,
				// ListID will be set by the caller
			}
			rules = append(rules, rule)
		} else {
			// Log warning for unrecognized format
			log.Printf("Warning: Unrecognized hosts format in line: %s", line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading hosts file: %w", err)
	}

	return rules, nil
}

func readDomainsFile(r io.Reader) ([]model.Rule, error) {
	var rules []model.Rule
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		line = whitespacePattern.ReplaceAllString(line, " ") // normalize whitespace
		line = trimPattern.ReplaceAllString(line, "")        // trim

		// Skip empty lines and comments
		if line == "" || line[0] == '#' {
			continue
		}

		// Validate domain format (line should be just a domain)
		if !validDomainPattern.MatchString(line) {
			log.Printf("Warning: Invalid domain format in domains file: %s", line)
			continue
		}

		// Create rule (default to blocking for domains file)
		rule := model.Rule{
			Domain:  line,
			Allowed: false,
			// ListID will be set by the caller
		}
		rules = append(rules, rule)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading domains file: %w", err)
	}

	return rules, nil
}

// detectAndReadFile auto-detects file format and calls appropriate parser
func detectAndReadFile(r io.Reader) ([]model.Rule, error) {
	// Read entire file content
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("error reading file for format detection: %w", err)
	}

	contentStr := string(content)

	if len(contentStr) == 0 {
		log.Printf("Warning: File appears to be empty")
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
		log.Printf("detected adblock format file")
		return readAdblockFile(strings.NewReader(contentStr))
	} else if hostsCount > 0 {
		log.Printf("detected hosts format file")
		return readHostFile(strings.NewReader(contentStr))
	} else if domainCount > 0 {
		log.Printf("detected domains format file")
		return readDomainsFile(strings.NewReader(contentStr))
	} else {
		log.Printf("warning: Unable to detect file format - no recognizable patterns found in first %d lines", sampleSize)
		return []model.Rule{}, nil
	}
}

// readHTTPSFile reads rules from an HTTPS file.
func readHTTPSFile(url string) ([]model.Rule, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to GET url %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}

	return detectAndReadFile(resp.Body)
}

// readLocalFile reads rules from a local file.
func readLocalFile(path string) ([]model.Rule, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open local file %s: %w", path, err)
	}
	defer file.Close()

	return detectAndReadFile(file)
}

func readFromSource(source string) ([]model.Rule, error) {
	switch {
	case httpsPattern.MatchString(source):
		return readHTTPSFile(source)
	case localFilePattern.MatchString(source):
		return readLocalFile(source)
	default:
		return nil, fmt.Errorf("neither HTTPS file, nor local file")
	}
}

func (ls *listsService) LoadList(l *model.List) error {
	// Check if no source to load from
	if l.Source == "" {
		return nil
	}

	// Read rules from the source
	rules, err := readFromSource(l.Source)
	if err != nil {
		return fmt.Errorf("failed to read from source %s: %w", l.Source, err)
	}

	// Save each rule to database with correct ListID
	for _, rule := range rules {
		rule.ListID = l.ID
		_, err := ls.ruleRepo.Create(&rule)
		if err != nil {
			return fmt.Errorf("failed to create rule for domain %s: %w", rule.Domain, err)
		}
	}

	// Load the updated rules into the list model
	l.Rules, err = ls.listRepo.LoadRules(l.ID)
	if err != nil {
		return fmt.Errorf("failed to load rules for list %d: %w", l.ID, err)
	}

	log.Printf("Successfully loaded %d rules for list %s from %s", len(l.Rules), l.Name, l.Source)
	return nil
}
