package utils

import (
	"strings"
)

const MultilineResponseDelimiter = "<br>"

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

// Helper function to send multiline responses to clients.
// This function replaces newline characters with '<br>' as the client expects a single line response.
func FormatMultilineResponse(response string) string {
	return strings.ReplaceAll(response, "\n", MultilineResponseDelimiter)
}

// Helper function to convert a multiline response back to its original form.
func ParseServerResponse(response string) string {
	response = strings.TrimSpace(response)
	// check if the response contains <br> and replace it with newline
	if strings.Contains(response, MultilineResponseDelimiter) {
		response = strings.ReplaceAll(response, MultilineResponseDelimiter, "\n")
		// add a newline at the start of the response for better formatting
		response = "\n" + response
	}
	return response
}
