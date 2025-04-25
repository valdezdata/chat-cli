package tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/valdezdata/chat-cli/internal/cli"
	"github.com/valdezdata/chat-cli/internal/logging"
	"github.com/valdezdata/chat-cli/internal/providers"
)

func TestCreateChatClient(t *testing.T) {
	os.Unsetenv("TOGETHER_API_KEY")
	os.Unsetenv("GROQ_API_KEY")
	os.Unsetenv("SAMBA_API_KEY")
	os.Unsetenv("GEMINI_API_KEY") // Add this line for Gemini testing

	// Create a logger for testing
	logger := logging.New()

	tests := []struct {
		name         string
		provider     cli.Provider
		expectError  bool
		expectedType string
	}{
		{
			name:         "Ollama no API key",
			provider:     cli.ProviderOllama,
			expectError:  false,
			expectedType: "*providers.OllamaClient",
		},
		{
			name:         "Groq without API key",
			provider:     cli.ProviderGroq,
			expectError:  true,
			expectedType: "",
		},
		{
			name:         "Gemini without API key",
			provider:     cli.ProviderGemini,
			expectError:  true,
			expectedType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Add logger parameter here
			client, err := cli.CreateChatClient(tt.provider, logger)
			if tt.expectError {
				if err == nil {
					t.Errorf("CreateChatClient(%q) expected error, got nil", tt.provider)
				}
				return
			}

			if err != nil {
				t.Errorf("CreateChatClient(%q) unexpected error: %v", tt.provider, err)
				return
			}

			// Keep providers import active
			var _ providers.ChatInterface = client

			gotType := fmt.Sprintf("%T", client)
			if gotType != tt.expectedType {
				t.Errorf("CreateChatClient(%q) type = %v, want %v", tt.provider, gotType, tt.expectedType)
			}
		})
	}
}
