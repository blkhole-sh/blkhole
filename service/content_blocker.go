package services

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"server/util"
	"strings"

	"github.com/armon/go-radix"
)

var domainRegex = regexp.MustCompile(`^(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

type ContentBlocker struct {
	tree *radix.Tree
}

func NewContentBlocker(blocklists ...string) (*ContentBlocker, error) {
	// Initialize new radix tree
	tree := radix.New()

	// Iterate over each blocklist
	for _, blocklist := range blocklists {

		// Try to open the blocklist file
		file, err := os.Open(blocklist)
		if err != nil {
			return nil, err
		}

		defer file.Close()

		// Initialize new file scanner
		scanner := bufio.NewScanner(file)

		// Iterate over each line in blocklist and parse domain
		for scanner.Scan() {
			domain := scanner.Text()

			// Check if domain is valid
			if !domainRegex.MatchString(domain) {
				return nil, fmt.Errorf("%s is not a valid domain\n", domain)
			}

			// Insert domain revered into radix tree
			tree.Insert(util.Reverse(domain), true)
		}
	}

	return &ContentBlocker{tree: tree}, nil
}

func (c *ContentBlocker) IsBlocked(domain string) bool {
	longestPrefix, _, found := c.tree.LongestPrefix(util.Reverse(domain))

	if found && len(strings.Split(longestPrefix, ".")) > 1 {
		return true
	}

	return false
}
