package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/valdezdata/chat-cli/internal/consts"
	"github.com/valdezdata/chat-cli/internal/logging" // Import logging
	"github.com/valdezdata/chat-cli/internal/utils"   // For RedactAPIKey

	"github.com/fatih/color"
)

// Supported SambaNova models.
var sambaModels = map[string]string{
	"llama-70b": "Meta-Llama-3.3-70B-Instruct",
	// Add other models if available
}

// SambaClient handles communication with the SambaNova API.
type SambaClient struct {
	apiKey        string
	baseURL       string
	messages      []Message // Using a local Message type for SambaNova
	selectedModel string
	httpClient    *http.Client
	logger        *logging.Logger // Logger instance
}

// Message represents a chat message for SambaNova API structure.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest for SambaNova API.
type ChatCompletionRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature"`
	Stream      bool      `json:"stream"` // Note: SambaNova implementation uses non-streaming
}

// ChatCompletionResponse from SambaNova API.
type ChatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// SetLogger injects the logger.
func (s *SambaClient) SetLogger(logger *logging.Logger) {
	s.logger = logger
}

// log helper for internal logging.
func (s *SambaClient) log(level logging.LogLevel, format string, args ...interface{}) {
	if s.logger != nil {
		switch level {
		case logging.DEBUG:
			s.logger.Debug(format, args...)
		case logging.INFO:
			s.logger.Info(format, args...)
		case logging.WARN:
			s.logger.Warn(format, args...)
		case logging.ERROR:
			s.logger.Error(format, args...)
		}
	}
}

// Initialize sets up the client with API key, base URL, and model.
func (s *SambaClient) Initialize() error {
	s.apiKey = os.Getenv("SAMBA_API_KEY")
	if s.apiKey == "" {
		return fmt.Errorf("SAMBA_API_KEY env var not set")
	}
	// Optional: Add API Key validation if SambaNova has specific rules
	// if !utils.ValidateAPIKey(s.apiKey, "samba") { ... }

	s.log(logging.DEBUG, "Initializing SambaNova client...")
	s.log(logging.DEBUG, "Using API key: %s", utils.RedactAPIKey(s.apiKey))

	modelEnv := os.Getenv("SAMBA_MODEL")
	s.selectedModel = sambaModels[modelEnv]
	if s.selectedModel == "" {
		defaultModel := sambaModels["llama-70b"]
		s.log(logging.WARN, "SAMBA_MODEL '%s' not found or not set, using default '%s'", modelEnv, defaultModel)
		s.selectedModel = defaultModel
	} else {
		s.log(logging.DEBUG, "Selected SambaNova model: %s", s.selectedModel)
	}

	s.baseURL = "https://api.sambanova.ai/v1"              // SambaNova API base URL
	s.httpClient = &http.Client{Timeout: 60 * time.Second} // Increased timeout for potentially slower models

	// Initial message list is empty for SambaNova, build context per request maybe?
	// Or initialize with system prompt if supported:
	// s.messages = []Message{{Role: consts.SystemRole, Content: "..."}}

	s.log(logging.INFO, "SambaNova client initialized (Model: %s)", s.selectedModel)
	return nil
}

// GetModelName returns the selected model identifier.
func (s *SambaClient) GetModelName() string {
	return s.selectedModel
}

// SendMessage sends a user message and gets a non-streamed response.
func (s *SambaClient) SendMessage(message string) (string, time.Duration, error) {
	s.log(logging.DEBUG, "SambaNova: Preparing message (%d chars)", len(message))

	// Append the new user message to the current context for this request
	// Note: SambaNova might require managing context differently than OpenAI's message list.
	// This assumes a simple append works. Adjust if state needs to be reset or managed.
	currentMessages := append(s.messages, Message{
		Role: consts.UserRole, Content: message,
	})

	start := time.Now()

	// Create request with context for timeout
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second) // Longer timeout for request
	defer cancel()

	req := ChatCompletionRequest{
		Model:       s.selectedModel,
		Messages:    currentMessages, // Send current conversation context
		Temperature: 0.7,
		Stream:      false, // SambaNova implementation uses non-streaming
	}

	s.log(logging.DEBUG, "SambaNova: Marshalling request payload")
	requestBody, err := json.Marshal(req)
	if err != nil {
		s.log(logging.ERROR, "SambaNova: Failed to marshal request: %v", err)
		return "", 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	s.log(logging.DEBUG, "SambaNova: Creating HTTP POST request to %s/chat/completions", s.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		s.log(logging.ERROR, "SambaNova: Failed to create HTTP request: %v", err)
		return "", 0, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+s.apiKey)

	s.log(logging.DEBUG, "SambaNova: Sending request...")
	resp, err := s.httpClient.Do(httpReq)
	if err != nil {
		// Handle context deadline exceeded specifically
		if ctx.Err() == context.DeadlineExceeded {
			s.log(logging.ERROR, "SambaNova: Request timed out: %v", err)
			return "", 0, fmt.Errorf("request timed out after 90s")
		}
		s.log(logging.ERROR, "SambaNova: Failed to send request: %v", err)
		return "", 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	s.log(logging.DEBUG, "SambaNova: Received response (Status: %s)", resp.Status)
	respBody, err := io.ReadAll(resp.Body) // Read body for potential error details
	if err != nil {
		s.log(logging.ERROR, "SambaNova: Failed to read response body: %v", err)
		// Still try to proceed if status code is OK, but log the read error
	}

	if resp.StatusCode != http.StatusOK {
		s.log(logging.ERROR, "SambaNova: API request failed (Status: %d): %s", resp.StatusCode, string(respBody))
		return "", 0, fmt.Errorf("API request failed (%d): %s", resp.StatusCode, string(respBody))
	}

	s.log(logging.DEBUG, "SambaNova: Decoding JSON response")
	var completionResp ChatCompletionResponse
	if err := json.Unmarshal(respBody, &completionResp); err != nil {
		s.log(logging.ERROR, "SambaNova: Failed to decode response JSON: %v", err)
		s.log(logging.DEBUG, "SambaNova: Raw response body: %s", string(respBody))
		return "", 0, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(completionResp.Choices) == 0 || completionResp.Choices[0].Message.Content == "" {
		s.log(logging.WARN, "SambaNova: No choices or empty content returned in response")
		// Return empty string but no error, as the API call succeeded technically
		return "", time.Since(start), nil
	}

	assistantColor := color.New(color.FgHiMagenta)
	assistantPrompt := assistantColor.PrintfFunc()
	assistantPrompt("%s Assistant: ", consts.RobotEmoji)

	content := completionResp.Choices[0].Message.Content
	elapsed := time.Since(start)
	s.log(logging.DEBUG, "SambaNova: Response received (%d chars) in %v", len(content), elapsed)

	assistantPrompt("%s\n", content) // Print the full response at once

	// Update the persistent message history for the next turn
	s.messages = append(currentMessages, Message{
		Role: consts.AssistantRole, Content: content,
	})

	return content, elapsed, nil
}
