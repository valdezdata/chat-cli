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
	"github.com/sashabaranov/go-openai" // Using openai client for Groq compatibility
)

// Supported Groq models.
var groqModels = map[string]string{
	"gemma": "gemma2-9b-it",
}

// GroqClient handles communication with the Groq API.
type GroqClient struct {
	client        *openai.Client
	messages      []openai.ChatCompletionMessage
	selectedModel string
	logger        *logging.Logger // Logger instance
}

// SetLogger injects the logger.
func (g *GroqClient) SetLogger(logger *logging.Logger) {
	g.logger = logger
}

// log helper for internal logging.
func (g *GroqClient) log(level logging.LogLevel, format string, args ...interface{}) {
	if g.logger != nil {
		switch level {
		case logging.DEBUG:
			g.logger.Debug(format, args...)
		case logging.INFO:
			g.logger.Info(format, args...)
		case logging.WARN:
			g.logger.Warn(format, args...)
		case logging.ERROR:
			g.logger.Error(format, args...)
		}
	}
}

// Initialize sets up the client with API key, base URL, and model.
func (g *GroqClient) Initialize() error {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("GROQ_API_KEY env var not set")
	}
	// Optional: Add API Key validation if Groq has specific rules
	// if !utils.ValidateAPIKey(apiKey, "groq") { ... }

	g.log(logging.DEBUG, "Initializing Groq client...")
	g.log(logging.DEBUG, "Using API key: %s", utils.RedactAPIKey(apiKey))

	modelEnv := os.Getenv("GROQ_MODEL")
	g.selectedModel = groqModels[modelEnv]
	if g.selectedModel == "" {
		defaultModel := groqModels["gemma"]
		g.log(logging.WARN, "GROQ_MODEL '%s' not found or not set, using default '%s'", modelEnv, defaultModel)
		g.selectedModel = defaultModel
	} else {
		g.log(logging.DEBUG, "Selected Groq model: %s", g.selectedModel)
	}

	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.groq.com/openai/v1" // Set Groq base URL
	g.client = openai.NewClientWithConfig(config)

	g.messages = []openai.ChatCompletionMessage{
		{Role: consts.SystemRole, Content: "Provide helpful and concise responses"},
	}

	g.log(logging.INFO, "Groq client initialized (Model: %s)", g.selectedModel)
	return nil
}

// GetModelName returns the selected model identifier.
func (g *GroqClient) GetModelName() string {
	return g.selectedModel
}

// SendMessage sends a user message and streams the response.
func (g *GroqClient) SendMessage(message string) (string, time.Duration, error) {
	g.log(logging.DEBUG, "Groq: Appending user message (%d chars)", len(message))
	g.messages = append(g.messages, openai.ChatCompletionMessage{
		Role: consts.UserRole, Content: message,
	})

	start := time.Now()
	req := openai.ChatCompletionRequest{
		Model:    g.selectedModel,
		Messages: g.messages,
		// Temperature: 0.7, // Add if needed
		Stream: true,
	}

	g.log(logging.DEBUG, "Groq: Creating stream for model %s", g.selectedModel)
	stream, err := g.client.CreateChatCompletionStream(context.Background(), req)
	if err != nil {
		g.log(logging.ERROR, "Groq: Failed to create stream: %v", err)
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
	g.log(logging.DEBUG, "Groq: Receiving stream...")
	for {
		response, err := stream.Recv()
		if err != nil {
			if err == io.EOF || strings.Contains(err.Error(), "EOF") {
				g.log(logging.DEBUG, "Groq: Stream finished (EOF)")
				break
			}
			g.log(logging.ERROR, "Groq: Stream receive error: %v", err)
			return fullResponse.String(), time.Since(start), fmt.Errorf("stream error: %w", err)
		}

		if len(response.Choices) > 0 {
			contentChunk := response.Choices[0].Delta.Content
			assistantPrompt("%s", contentChunk) // Stream output
			fullResponse.WriteString(contentChunk)
		} else {
			g.log(logging.WARN, "Groq: Received stream response with no choices")
		}
	}
	fmt.Println() // Newline after streaming

	elapsed := time.Since(start)
	finalResponseStr := fullResponse.String()
	g.log(logging.DEBUG, "Groq: Response received (%d chars) in %v", len(finalResponseStr), elapsed)

	g.messages = append(g.messages, openai.ChatCompletionMessage{
		Role: consts.AssistantRole, Content: finalResponseStr,
	})

	return finalResponseStr, elapsed, nil
}
