package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/valdezdata/chat-cli/internal/consts"
	"github.com/valdezdata/chat-cli/internal/logging" // Import logging

	"github.com/fatih/color"
)

// Supported Ollama models (local names).
var ollamaModels = map[string]string{
	"mistral":  "mistral:latest",
	"llama":    "llama3:latest",
	"deepseek": "deepseek-coder:latest",
	"gemma":    "gemma:latest",
}

// OllamaMessage structure for Ollama API.
type OllamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OllamaRequest structure for Ollama API.
type OllamaRequest struct {
	Model    string          `json:"model"`
	Messages []OllamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
	// Options map[string]interface{} `json:"options,omitempty"` // Add for temp, etc.
}

// OllamaResponse structure for Ollama stream chunks.
type OllamaResponse struct {
	Model     string        `json:"model"`
	CreatedAt time.Time     `json:"created_at"`
	Message   OllamaMessage `json:"message"`
	Done      bool          `json:"done"`
	// Other fields like total_duration, etc., might be present when done=true
}

// OllamaClient handles communication with a local Ollama instance.
type OllamaClient struct {
	serverURL     string
	messages      []OllamaMessage
	selectedModel string
	httpClient    *http.Client    // Use a shared client
	logger        *logging.Logger // Logger instance
}

// SetLogger injects the logger.
func (o *OllamaClient) SetLogger(logger *logging.Logger) {
	o.logger = logger
}

// log helper for internal logging.
func (o *OllamaClient) log(level logging.LogLevel, format string, args ...interface{}) {
	if o.logger != nil {
		switch level {
		case logging.DEBUG:
			o.logger.Debug(format, args...)
		case logging.INFO:
			o.logger.Info(format, args...)
		case logging.WARN:
			o.logger.Warn(format, args...)
		case logging.ERROR:
			o.logger.Error(format, args...)
		}
	}
}

// Initialize sets up the client, checks connection, and selects model.
func (o *OllamaClient) Initialize() error {
	o.serverURL = os.Getenv("OLLAMA_URL")
	if o.serverURL == "" {
		o.serverURL = "http://localhost:11434" // Default local URL
	}
	o.httpClient = &http.Client{Timeout: 60 * time.Second} // Client with timeout

	o.log(logging.DEBUG, "Initializing Ollama client (URL: %s)...", o.serverURL)

	modelEnv := os.Getenv("OLLAMA_MODEL")
	o.selectedModel = ollamaModels[modelEnv]
	if o.selectedModel == "" {
		defaultModel := ollamaModels["llama"]
		o.log(logging.WARN, "OLLAMA_MODEL '%s' not found or not set, using default '%s'", modelEnv, defaultModel)
		o.selectedModel = defaultModel
	} else {
		o.log(logging.DEBUG, "Selected Ollama model: %s", o.selectedModel)
	}

	o.messages = []OllamaMessage{
		{Role: consts.SystemRole, Content: "Provide helpful and concise responses"},
	}

	// Check connection to Ollama server
	o.log(logging.DEBUG, "Checking connection to Ollama server at %s", o.serverURL)
	resp, err := o.httpClient.Get(o.serverURL) // Simple GET to base URL often works
	if err != nil {
		o.log(logging.ERROR, "Could not connect to Ollama server at %s: %v", o.serverURL, err)
		return fmt.Errorf("could not connect to Ollama server at %s: %w", o.serverURL, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		o.log(logging.ERROR, "Ollama server connection check failed (Status: %d): %s", resp.StatusCode, string(bodyBytes))
		return fmt.Errorf("ollama server at %s returned status %d", o.serverURL, resp.StatusCode)
	}
	o.log(logging.DEBUG, "Ollama server connection successful.")

	// Optional: Check if the selected model exists using /api/tags
	// ... (implementation for checking model existence) ...

	o.log(logging.INFO, "Ollama client initialized (URL: %s, Model: %s)", o.serverURL, o.selectedModel)
	return nil
}

// GetModelName returns the selected model identifier.
func (o *OllamaClient) GetModelName() string {
	return o.selectedModel
}

// SendMessage sends a user message and streams the response from Ollama.
func (o *OllamaClient) SendMessage(message string) (string, time.Duration, error) {
	o.log(logging.DEBUG, "Ollama: Appending user message (%d chars)", len(message))
	o.messages = append(o.messages, OllamaMessage{
		Role: consts.UserRole, Content: message,
	})

	start := time.Now()

	reqPayload := OllamaRequest{
		Model:    o.selectedModel,
		Messages: o.messages,
		Stream:   true,
		// Options: map[string]interface{}{"temperature": 0.7}, // Example options
	}

	reqData, err := json.Marshal(reqPayload)
	if err != nil {
		o.log(logging.ERROR, "Ollama: Failed to marshal request: %v", err)
		return "", 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	o.log(logging.DEBUG, "Ollama: Sending request to %s/api/chat for model %s", o.serverURL, o.selectedModel)
	resp, err := o.httpClient.Post(o.serverURL+"/api/chat", "application/json", bytes.NewBuffer(reqData))
	if err != nil {
		o.log(logging.ERROR, "Ollama: Failed to send request: %v", err)
		return "", 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		o.log(logging.ERROR, "Ollama: API request failed (Status: %d): %s", resp.StatusCode, string(bodyBytes))
		return "", 0, fmt.Errorf("API request failed (%d): %s", resp.StatusCode, string(bodyBytes))
	}

	assistantColor := color.New(color.FgHiMagenta)
	assistantPrompt := assistantColor.PrintfFunc()
	assistantPrompt("%s Assistant: ", consts.RobotEmoji) // Prefix only once

	var fullResponse strings.Builder
	decoder := json.NewDecoder(resp.Body)

	o.log(logging.DEBUG, "Ollama: Receiving stream...")
	for {
		var ollamaResp OllamaResponse
		if err := decoder.Decode(&ollamaResp); err != nil {
			if err == io.EOF {
				o.log(logging.DEBUG, "Ollama: Stream finished (EOF)")
				break // End of stream
			}
			o.log(logging.ERROR, "Ollama: Failed to decode stream chunk: %v", err)
			// Return partial response with error
			return fullResponse.String(), time.Since(start), fmt.Errorf("failed to decode response: %w", err)
		}

		contentChunk := ollamaResp.Message.Content
		assistantPrompt("%s", contentChunk) // Stream output
		fullResponse.WriteString(contentChunk)

		// Check the 'done' field which Ollama sends in the last chunk
		if ollamaResp.Done {
			o.log(logging.DEBUG, "Ollama: Received 'done' flag in stream.")
			break
		}
	}
	fmt.Println() // Newline after streaming

	elapsed := time.Since(start)
	finalResponseStr := fullResponse.String()
	o.log(logging.DEBUG, "Ollama: Response received (%d chars) in %v", len(finalResponseStr), elapsed)

	// Append the full response to maintain conversation context
	o.messages = append(o.messages, OllamaMessage{
		Role: consts.AssistantRole, Content: finalResponseStr,
	})

	return finalResponseStr, elapsed, nil
}
