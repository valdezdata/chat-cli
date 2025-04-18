package providers

import (
	"context"
	"fmt"
	"io" // Needed for io.EOF check
	"os"
	"strings"
	"time"

	"github.com/valdezdata/chat-cli/internal/consts"
	"github.com/valdezdata/chat-cli/internal/logging" // Import logging
	"github.com/valdezdata/chat-cli/internal/utils"   // For RedactAPIKey

	"github.com/fatih/color"
	"github.com/sashabaranov/go-openai"
)

// Supported Together AI models.
var togetherModels = map[string]string{
	"llama-70b": "meta-llama/Llama-3.3-70B-Instruct-Turbo-Free",
	"deepseek":  "deepseek-ai/DeepSeek-R1-Distill-Llama-70B-free",
}

// TogetherClient handles communication with the Together AI API.
type TogetherClient struct {
	client        *openai.Client
	messages      []openai.ChatCompletionMessage
	selectedModel string
	logger        *logging.Logger // Logger instance
}

// SetLogger injects the logger.
func (t *TogetherClient) SetLogger(logger *logging.Logger) {
	t.logger = logger
}

// log helper for internal logging.
func (t *TogetherClient) log(level logging.LogLevel, format string, args ...interface{}) {
	if t.logger != nil {
		switch level {
		case logging.DEBUG:
			t.logger.Debug(format, args...)
		case logging.INFO:
			t.logger.Info(format, args...)
		case logging.WARN:
			t.logger.Warn(format, args...)
		case logging.ERROR:
			t.logger.Error(format, args...)
		}
	}
}

// Initialize sets up the client with API key, base URL, and model.
func (t *TogetherClient) Initialize() error {
	apiKey := os.Getenv("TOGETHER_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("TOGETHER_API_KEY env var not set")
	}
	// Optional: Add API Key validation if Together has specific rules
	// if !utils.ValidateAPIKey(apiKey, "together") { ... }

	t.log(logging.DEBUG, "Initializing Together client...")
	t.log(logging.DEBUG, "Using API key: %s", utils.RedactAPIKey(apiKey))

	modelEnv := os.Getenv("TOGETHER_MODEL")
	t.selectedModel = togetherModels[modelEnv]
	if t.selectedModel == "" {
		defaultModel := togetherModels["llama-70b"]
		t.log(logging.WARN, "TOGETHER_MODEL '%s' not found or not set, using default '%s'", modelEnv, defaultModel)
		t.selectedModel = defaultModel
	} else {
		t.log(logging.DEBUG, "Selected Together model: %s", t.selectedModel)
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.together.xyz/v1" // Set Together base URL
	t.client = openai.NewClientWithConfig(config)

	t.messages = []openai.ChatCompletionMessage{
		{Role: consts.SystemRole, Content: "Provide helpful and concise responses"},
	}

	t.log(logging.INFO, "Together client initialized (Model: %s)", t.selectedModel)
	return nil
}

// GetModelName returns the selected model identifier.
func (t *TogetherClient) GetModelName() string {
	return t.selectedModel
}

// SendMessage sends a user message and streams the response.
func (t *TogetherClient) SendMessage(message string) (string, time.Duration, error) {
	t.log(logging.DEBUG, "Together: Appending user message (%d chars)", len(message))
	t.messages = append(t.messages, openai.ChatCompletionMessage{
		Role: consts.UserRole, Content: message,
	})

	start := time.Now()
	req := openai.ChatCompletionRequest{
		Model:    t.selectedModel,
		Messages: t.messages,
		// Temperature: 0.7, // Add if needed
		Stream: true,
	}

	t.log(logging.DEBUG, "Together: Creating stream for model %s", t.selectedModel)
	stream, err := t.client.CreateChatCompletionStream(context.Background(), req)
	if err != nil {
		t.log(logging.ERROR, "Together: Failed to create stream: %v", err)
		errMsg := fmt.Sprintf("failed to create stream: %v", err)
		// Check for specific API errors if the library provides them
		// if apiErr, ok := err.(*openai.APIError); ok { ... }
		return "", 0, fmt.Errorf(errMsg)
	}
	defer stream.Close()

	assistantColor := color.New(color.FgHiMagenta)
	assistantPrompt := assistantColor.PrintfFunc()
	assistantPrompt("%s Assistant: ", consts.RobotEmoji) // Prefix only once

	var fullResponse strings.Builder
	t.log(logging.DEBUG, "Together: Receiving stream...")
	for {
		response, err := stream.Recv()
		if err != nil {
			if err == io.EOF || strings.Contains(err.Error(), "EOF") {
				t.log(logging.DEBUG, "Together: Stream finished (EOF)")
				break
			}
			t.log(logging.ERROR, "Together: Stream receive error: %v", err)
			return fullResponse.String(), time.Since(start), fmt.Errorf("stream error: %w", err)
		}

		if len(response.Choices) > 0 {
			contentChunk := response.Choices[0].Delta.Content
			assistantPrompt("%s", contentChunk) // Stream output
			fullResponse.WriteString(contentChunk)
		} else {
			t.log(logging.WARN, "Together: Received stream response with no choices")
		}
	}
	fmt.Println() // Newline after streaming

	elapsed := time.Since(start)
	finalResponseStr := fullResponse.String()
	t.log(logging.DEBUG, "Together: Response received (%d chars) in %v", len(finalResponseStr), elapsed)

	t.messages = append(t.messages, openai.ChatCompletionMessage{
		Role: consts.AssistantRole, Content: finalResponseStr,
	})

	return finalResponseStr, elapsed, nil
}
