package utils

// RedactAPIKey returns a safely redacted version of an API key for logging
func RedactAPIKey(key string) string {
	if len(key) < 8 {
		return "[REDACTED_KEY]"
	}

	// Show first 4 and last 4 characters, mask the rest
	redacted := key[:4] + "..." + key[len(key)-4:]
	return redacted
}
