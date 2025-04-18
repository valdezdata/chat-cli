package providers

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/valdezdata/chat-cli/internal/consts"
	"github.com/valdezdata/chat-cli/internal/logging"
	"github.com/valdezdata/chat-cli/internal/utils"

	"github.com/fatih/color"
	"github.com/sashabaranov/go-openai"
)

// Supported OpenAI models.
var openaiModels = map[string]string{
	"gpt-4.1-nano": "gpt-4.1-nano",
}

// OpenAIClient handles communication with the OpenAI API.
type OpenAIClient struct {
	client        *openai.Client
	messages      []openai.ChatCompletionMessage
	selectedModel string
	logger        *logging.Logger // Logger instance
}

// SetLogger injects the logger.
func (o *OpenAIClient) SetLogger(logger *logging.Logger) {
	o.logger = logger
}

// log helper for internal logging.
func (o *OpenAIClient) log(level logging.LogLevel, format string, args ...interface{}) {
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

// Initialize sets up the client with API key and model.
func (o *OpenAIClient) Initialize() error {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OPENAI_API_KEY env var not set")
	}

	if !utils.ValidateAPIKey(apiKey, "openai") {
		return fmt.Errorf("OPENAI_API_KEY appears invalid")
	}

	o.log(logging.DEBUG, "Initializing OpenAI client...")
	o.log(logging.DEBUG, "Using API key: %s", utils.RedactAPIKey(apiKey))

	o.client = openai.NewClient(apiKey)

	modelEnv := os.Getenv("OPENAI_MODEL")
	o.selectedModel = openaiModels[modelEnv]
	if o.selectedModel == "" {
		defaultModel := openaiModels["gpt-4.1-nano"]
		o.log(logging.WARN, "OPENAI_MODEL '%s' not found or not set, using default '%s'", modelEnv, defaultModel)
		o.selectedModel = defaultModel
	} else {
		o.log(logging.DEBUG, "Selected OpenAI model: %s", o.selectedModel)
	}

	o.messages = []openai.ChatCompletionMessage{
		{Role: consts.SystemRole, Content: "Provide helpful and concise responses"},
	}

	o.log(logging.INFO, "OpenAI client initialized (Model: %s)", o.selectedModel)
	return nil
}

// GetModelName returns the selected model identifier.
func (o *OpenAIClient) GetModelName() string {
	return o.selectedModel
}

// SendMessage sends a user message and streams the response.
func (o *OpenAIClient) SendMessage(message string) (string, time.Duration, error) {
	o.log(logging.DEBUG, "OpenAI: Appending user message (%d chars)", len(message))
	o.messages = append(o.messages, openai.ChatCompletionMessage{
		Role: consts.UserRole, Content: message,
	})

	start := time.Now()
	req := openai.ChatCompletionRequest{
		Model:    o.selectedModel,
		Messages: o.messages,
		// Temperature: 0.7, // Add if needed, or get from opts
		Stream: true,
	}

	o.log(logging.DEBUG, "OpenAI: Creating stream for model %s", o.selectedModel)
	stream, err := o.client.CreateChatCompletionStream(context.Background(), req)
	if err != nil {
		o.log(logging.ERROR, "OpenAI: Failed to create stream: %v", err)
		errMsg := fmt.Sprintf("failed to create stream: %v", err)
		if apiErr, ok := err.(*openai.APIError); ok {
			errMsg = fmt.Sprintf("OpenAI API error (%d): %s - %s", apiErr.HTTPStatusCode, apiErr.Code, apiErr.Message)
		}
		return "", 0, fmt.Errorf(errMsg)
	}
	defer stream.Close()

	assistantColor := color.New(color.FgHiMagenta)
	assistantPrompt := assistantColor.PrintfFunc()
	assistantPrompt("%s Assistant: ", consts.RobotEmoji) // Prefix only once

	var fullResponse strings.Builder
	o.log(logging.DEBUG, "OpenAI: Receiving stream...")
	for {
		response, err := stream.Recv()
		if err != nil {
			if err == io.EOF || strings.Contains(err.Error(), "EOF") {
				o.log(logging.DEBUG, "OpenAI: Stream finished (EOF)")
				break
			}
			o.log(logging.ERROR, "OpenAI: Stream receive error: %v", err)
			return fullResponse.String(), time.Since(start), fmt.Errorf("stream error: %w", err)
		}

		if len(response.Choices) > 0 {
			contentChunk := response.Choices[0].Delta.Content
			assistantPrompt("%s", contentChunk) // Stream output
			fullResponse.WriteString(contentChunk)
		} else {
			o.log(logging.WARN, "OpenAI: Received stream response with no choices")
		}
	}
	fmt.Println() // Newline after streaming

	elapsed := time.Since(start)
	finalResponseStr := fullResponse.String()
	o.log(logging.DEBUG, "OpenAI: Response received (%d chars) in %v", len(finalResponseStr), elapsed)

	o.messages = append(o.messages, openai.ChatCompletionMessage{
		Role: consts.AssistantRole, Content: finalResponseStr,
	})

	return finalResponseStr, elapsed, nil
}
