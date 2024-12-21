package utils

import "path/filepath"

// Helper function for pattern matching
func MatchPattern(key, pattern string) bool {
	// Simple glob matching; for more complex patterns, use regex if necessary in the future
	matched, _ := filepath.Match(pattern, key)
	return matched
}
