package utils

import (
	"unicode"
)

// ValidateAPIKey performs basic validation on API keys
func ValidateAPIKey(key string, provider string) bool {
	// Empty check (already done elsewhere, but for completeness)
	if key == "" {
		return false
	}

	// Basic length check - most API keys are at least 20 chars
	if len(key) < 20 {
		return false
	}

	// Character validation
	for _, c := range key {
		if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '-' && c != '_' && c != '.' {
			return false
		}
	}

	// Provider-specific checks
	switch provider {
	case "openai":
		// OpenAI keys often start with "sk-"
		if len(key) >= 3 && key[:3] != "sk-" {
			return false
		}
	}

	return true
}
