package validation

import (
	"regexp"
	"strings"
)

// NetID regex: 2-3 letters followed by 3 numbers
var netIDRegex = regexp.MustCompile(`^[a-zA-Z]{1,3}[0-9]{1,3}$`)

// IsValidNetID validates Rutgers NetID format
func IsValidNetID(netID string) bool {
	// Convert to lowercase and trim whitespace
	netID = strings.ToLower(strings.TrimSpace(netID))

	// Check regex pattern
	return netIDRegex.MatchString(netID)
}

// NormalizeNetID cleans up NetID input
func NormalizeNetID(netID string) string {
	return strings.ToLower(strings.TrimSpace(netID))
}

// Examples of valid NetIDs:
// - sab468 (3 letters + 3 numbers)
// - sb468  (2 letters + 3 numbers)
// - abc123 (3 letters + 3 numbers)
//
// Examples of invalid NetIDs:
// - s468     (only 1 letter)
// - sabe468  (4 letters)
// - sab12    (only 2 numbers)
// - sab1234  (4 numbers)
