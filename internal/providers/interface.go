package providers

import "time"

// MessageParams defines optional parameters for message sending
type MessageParams struct {
	Temperature float64
	MaxTokens   int
}

// ChatInterface defines the common interface for all chat providers
type ChatInterface interface {
	Initialize() error
	SendMessage(message string) (string, time.Duration, error)
	GetModelName() string

	// These can be implemented in a future update
	// SetTemperature(temp float64)
	// SetMaxTokens(tokens int)
}
