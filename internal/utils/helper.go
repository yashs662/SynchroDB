package utils

import (
	"strings"
)

// Helper function for pattern matching.
func MatchPattern(key, pattern string) bool {
	// If the pattern contains a wildcard '*', convert it to a regex-like pattern
	if strings.Contains(pattern, "*") {
		// Replace '*' with '.*' to match any sequence of characters
		regexPattern := strings.ReplaceAll(pattern, "*", ".*")
		return strings.Contains(key, strings.Trim(regexPattern, ".*"))
	}
	// Otherwise, perform a simple substring check
	return strings.Contains(key, pattern)
}
