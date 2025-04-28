package providers

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/valdezdata/chat-cli/internal/consts"
	"github.com/valdezdata/chat-cli/internal/logging"
	"github.com/valdezdata/chat-cli/internal/utils"

	"github.com/fatih/color"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Supported Gemini models.
var geminiModels = map[string]string{
	"gemini-pro":        "gemini-2.5-pro-exp-03-25",
	"gemini-flash":      "gemini-2.5-flash-preview-04-17",
	"gemini-flash-lite": "gemini-2.0-flash-lite",
}

// GeminiClient handles communication with the Google Gemini API.
type GeminiClient struct {
	client        *genai.Client
	model         *genai.GenerativeModel
	messages      []*genai.Content
	selectedModel string
	logger        *logging.Logger
}

// SetLogger injects the logger.
func (g *GeminiClient) SetLogger(logger *logging.Logger) {
	g.logger = logger
}

// log helper for internal logging.
func (g *GeminiClient) log(level logging.LogLevel, format string, args ...interface{}) {
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

// Initialize sets up the client with API key and model.
func (g *GeminiClient) Initialize() error {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("GEMINI_API_KEY env var not set")
	}

	g.log(logging.DEBUG, "Initializing Gemini client...")
	g.log(logging.DEBUG, "Using API key: %s", utils.RedactAPIKey(apiKey))

	// Set up the client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		g.log(logging.ERROR, "Failed to create Gemini client: %v", err)
		return fmt.Errorf("failed to create Gemini client: %w", err)
	}
	g.client = client

	// Get model from environment or use default
	modelEnv := os.Getenv("GEMINI_MODEL")
	g.selectedModel = geminiModels[modelEnv]
	if g.selectedModel == "" {
		defaultModel := geminiModels["gemini-flash-lite"]
		g.log(logging.WARN, "GEMINI_MODEL '%s' not found or not set, using default '%s'", modelEnv, defaultModel)
		g.selectedModel = defaultModel
	} else {
		g.log(logging.DEBUG, "Selected Gemini model: %s", g.selectedModel)
	}

	// Create the model and set parameters
	g.model = g.client.GenerativeModel(g.selectedModel)
	g.model.SetTemperature(0.7)
	g.model.SetTopP(0.95)

	// Initialize with empty conversation history
	// Note: Gemini doesn't support system role messages
	g.messages = []*genai.Content{}

	g.log(logging.INFO, "Gemini client initialized (Model: %s)", g.selectedModel)
	return nil
}

// GetModelName returns the selected model identifier.
func (g *GeminiClient) GetModelName() string {
	return g.selectedModel
}

// SendMessage sends a user message and streams the response.
func (g *GeminiClient) SendMessage(message string) (string, time.Duration, error) {
	g.log(logging.DEBUG, "Gemini: Processing user message (%d chars)", len(message))

	// Add user message to conversation history
	g.messages = append(g.messages, &genai.Content{
		Role: "user",
		Parts: []genai.Part{
			genai.Text(message),
		},
	})

	start := time.Now()

	// Create a chat session
	ctx := context.Background()
	chat := g.model.StartChat()

	// Only set history if we have previous messages
	if len(g.messages) > 1 { // More than just the current message
		chat.History = g.messages[:len(g.messages)-1] // Exclude the current message
	}

	// Send the message and stream the response
	g.log(logging.DEBUG, "Gemini: Sending request and starting stream")
	iter := chat.SendMessageStream(ctx, genai.Text(message))

	assistantColor := color.New(color.FgHiMagenta)
	assistantPrompt := assistantColor.PrintfFunc()
	assistantPrompt("%s Assistant: ", consts.RobotEmoji) // Prefix only once

	var fullResponse strings.Builder

	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			g.log(logging.DEBUG, "Gemini: Stream finished")
			break
		}
		if err != nil {
			g.log(logging.ERROR, "Gemini: Stream error: %v", err)
			return fullResponse.String(), time.Since(start), fmt.Errorf("stream error: %w", err)
		}

		// Display and collect response chunks
		for _, part := range resp.Candidates[0].Content.Parts {
			chunk := part.(genai.Text)
			chunkStr := string(chunk)
			assistantPrompt("%s", chunkStr) // Stream output
			fullResponse.WriteString(chunkStr)
		}
	}
	fmt.Println() // Newline after streaming

	elapsed := time.Since(start)
	finalResponseStr := fullResponse.String()
	g.log(logging.DEBUG, "Gemini: Response received (%d chars) in %v", len(finalResponseStr), elapsed)

	// Add assistant response to conversation history
	g.messages = append(g.messages, &genai.Content{
		Role: "assistant",
		Parts: []genai.Part{
			genai.Text(finalResponseStr),
		},
	})

	return finalResponseStr, elapsed, nil
}
