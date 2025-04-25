package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/valdezdata/chat-cli/internal/assessment"
	"github.com/valdezdata/chat-cli/internal/consts"
	"github.com/valdezdata/chat-cli/internal/history"
	"github.com/valdezdata/chat-cli/internal/logging"
	"github.com/valdezdata/chat-cli/internal/providers"

	"github.com/fatih/color"
)

type Provider string

const (
	ProviderTogether Provider = "together"
	ProviderOllama   Provider = "ollama"
	ProviderGroq     Provider = "groq"
	ProviderSamba    Provider = "samba"
	ProviderOpenAI   Provider = "openai"
	ProviderGemini   Provider = "gemini"
)

type ChatOptions struct {
	Verbose      bool
	Provider     Provider
	Assess       bool
	Shell        bool
	ShellPrompt  string
	Temperature  float64
	MaxTokens    int
	OutputFormat string
	LogLevel     string
	LogToFile    bool
	LogToConsole bool
	SkipHistory  bool
}

func setupLogging(opts *ChatOptions) (*logging.Logger, error) {
	config := logging.DefaultConfig()

	// Set log level based on command line flag
	switch opts.LogLevel {
	case "debug":
		config.Level = logging.DEBUG
	case "info":
		config.Level = logging.INFO
	case "warn":
		config.Level = logging.WARN
	case "error":
		config.Level = logging.ERROR
	}

	// Increase verbosity for debug level
	if opts.Verbose {
		config.Level = logging.DEBUG
	}

	// By default, don't log to console
	config.Console = opts.LogToConsole

	// Enable file logging if requested
	config.File = opts.LogToFile

	// Setup logger
	logger, err := logging.Setup(config)
	if err != nil {
		return nil, err
	}

	return logger, nil
}

type ProviderFlag Provider

func (p *ProviderFlag) String() string {
	return string(*p)
}

func (p *ProviderFlag) Set(value string) error {
	switch value {
	case string(ProviderTogether), string(ProviderOllama), string(ProviderGroq), string(ProviderSamba), string(ProviderOpenAI), string(ProviderGemini):
		*p = ProviderFlag(value)
		return nil
	default:
		return fmt.Errorf("must be one of: together, ollama, groq, samba, openai")
	}
}

func (p *ProviderFlag) Type() string {
	return "provider"
}

func clearScreen() {
	cmd := exec.Command("clear")
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// Function to format output based on requested format
func formatOutput(response, format string) string {
	switch format {
	case "json":
		// Simple JSON wrapping
		jsonResponse := fmt.Sprintf(`{"response": %q}`, response)
		return jsonResponse
	case "markdown":
		// Format as markdown
		return fmt.Sprintf("# LLM Response\n\n%s", response)
	default:
		// Default is plain text
		return response
	}
}

func CreateChatClient(provider Provider, logger *logging.Logger) (providers.ChatInterface, error) {
	logger.Debug("Creating chat client for provider: %s", provider)

	var client providers.ChatInterface
	var err error

	switch provider {
	case ProviderTogether:
		if os.Getenv("TOGETHER_API_KEY") != "" {
			logger.Debug("Initializing Together client")
			client = &providers.TogetherClient{}
			err = client.Initialize()
		} else {
			logger.Error("TOGETHER_API_KEY not set for Together provider")
			return nil, fmt.Errorf("TOGETHER_API_KEY not set for Together provider")
		}
	case ProviderGroq:
		if os.Getenv("GROQ_API_KEY") != "" {
			logger.Debug("Initializing Groq client")
			client = &providers.GroqClient{}
			err = client.Initialize()
		} else {
			logger.Error("GROQ_API_KEY not set for Groq provider")
			return nil, fmt.Errorf("GROQ_API_KEY not set for Groq provider")
		}
	case ProviderSamba:
		if os.Getenv("SAMBA_API_KEY") != "" {
			logger.Debug("Initializing Samba client")
			client = &providers.SambaClient{}
			err = client.Initialize()
		} else {
			logger.Error("SAMBA_API_KEY not set for Samba provider")
			return nil, fmt.Errorf("SAMBA_API_KEY not set for Samba provider")
		}
	case ProviderOpenAI:
		if os.Getenv("OPENAI_API_KEY") != "" {
			logger.Debug("Initializing OpenAI client")
			client = &providers.OpenAIClient{}
			err = client.Initialize()
		} else {
			logger.Error("OPENAI_API_KEY not set for OpenAI provider")
			return nil, fmt.Errorf("OPENAI_API_KEY not set for OpenAI provider")
		}
	case ProviderGemini:
		if os.Getenv("GEMINI_API_KEY") != "" {
			logger.Debug("Initializing Gemini client")
			client = &providers.GeminiClient{}
			err = client.Initialize()
		} else {
			logger.Error("GEMINI_API_KEY not set for Gemini provider")
			return nil, fmt.Errorf("GEMINI_API_KEY not set for Gemini provider")
		}
	case ProviderOllama:
		logger.Debug("Initializing Ollama client")
		client = &providers.OllamaClient{}
		err = client.Initialize()
	default:
		logger.Error("Unsupported provider: %s", provider)
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	if err != nil {
		logger.Error("Failed to initialize client: %v", err)
		return nil, err
	}

	// If the client supports setting a logger, set it
	if loggerAware, ok := client.(interface{ SetLogger(*logging.Logger) }); ok {
		logger.Debug("Setting logger for client")
		loggerAware.SetLogger(logger)
	}

	logger.Info("Successfully created chat client for %s provider", provider)
	return client, nil
}

// sendMessageAndLogHistory sends a message to the LLM and logs the interaction to history
func sendMessageAndLogHistory(client providers.ChatInterface, text string, opts *ChatOptions, logger *logging.Logger) (string, time.Duration, error) {
	// Send the message using the existing client
	response, elapsed, err := client.SendMessage(text)
	if err != nil {
		return "", elapsed, err
	}

	// Skip history if requested
	if opts.SkipHistory {
		logger.Debug("Skipping history logging as requested")
		return response, elapsed, nil
	}

	// Calculate approximate token counts (simple word-based approximation)
	inputTokens := len(strings.Split(text, " "))
	outputTokens := len(strings.Split(response, " "))
	totalTokens := inputTokens + outputTokens

	// Create history entry
	entry := history.Entry{
		Timestamp:    time.Now(),
		Provider:     string(opts.Provider),
		ModelName:    client.GetModelName(),
		Prompt:       text,
		Response:     response,
		InputTokens:  inputTokens,
		OutputTokens: outputTokens,
		TotalTokens:  totalTokens,
		TimeTaken:    elapsed.Seconds(),
	}

	// Add assessment if enabled
	if opts.Assess {
		// Create an assessment from the assessment package results
		promptAssessment := assessment.EvaluatePromptForHistory(text)

		// Convert to history assessment format
		assessmentEntry := &history.Assessment{
			OverallScore:   promptAssessment.TotalScore,
			OverallRating:  promptAssessment.OverallRating,
			CriteriaScores: make(map[string]history.CriteriaResult),
		}

		// Add each criteria
		for name, result := range promptAssessment.Criteria {
			assessmentEntry.CriteriaScores[name] = history.CriteriaResult{
				Score:       result.Score,
				Rating:      result.Rating,
				Description: result.Description,
			}
		}

		entry.Assessment = assessmentEntry
	}

	// Log history entry
	if err := history.AddEntry(entry); err != nil {
		logger.Error("Failed to log history: %v", err)
	}

	return response, elapsed, nil
}

func printMetrics(text, response string, elapsed time.Duration, assess bool) {
	metricsColor := color.New(color.FgHiYellow)

	inputTokens := len(strings.Split(text, " "))
	outputTokens := len(strings.Split(response, " "))
	totalTokens := inputTokens + outputTokens
	tokensPerSecond := float64(totalTokens) / elapsed.Seconds()

	metricsColor.Println("\nMetrics:")
	metricsColor.Printf("Time taken: %.2f seconds\n", elapsed.Seconds())
	metricsColor.Printf("Speed: %.2f tokens/second\n", tokensPerSecond)
	metricsColor.Printf("Input tokens: %d (approximate)\n", inputTokens)
	metricsColor.Printf("Output tokens: %d (approximate)\n", outputTokens)
	metricsColor.Printf("Total tokens: %d (approximate)\n", totalTokens)

	if assess {
		assessment.AssessPrompt(text)
	}
}

func handleUserInput(scanner *bufio.Scanner, userPrompt func(format string, a ...interface{})) (string, bool) {
	userPrompt("\n%s You: ", consts.PersonEmoji)
	scanner.Scan()
	text := scanner.Text()

	switch text {
	case "exit":
		return "", true
	case "clear":
		clearScreen()
		return "", false
	case "paste":
		userPrompt("Enter your text (type 'done' on a new line when finished):\n")
		var builder strings.Builder
		for scanner.Scan() {
			line := scanner.Text()
			if line == "done" {
				break
			}
			builder.WriteString(line + "\n")
		}
		return strings.TrimSpace(builder.String()), false
	default:
		return text, false
	}
}

func ShellMode(opts *ChatOptions, logger *logging.Logger) {
	logger.Debug("Initializing shell mode with options: %+v", opts)

	client, err := CreateChatClient(opts.Provider, logger)
	if err != nil {
		logger.Error("Failed to create chat client: %v", err)
		color.Red("Error: %v", err)
		return
	}

	// Check if there's input from stdin
	info, err := os.Stdin.Stat()
	if err != nil {
		logger.Error("Failed to get stdin info: %v", err)
		color.Red("Error: %v", err)
		return
	}

	// Read from stdin if there's piped input
	var input string
	var stdinContent string

	// Check if content is being piped in
	if (info.Mode() & os.ModeCharDevice) == 0 {
		logger.Debug("Detected piped input from stdin")
		// Read directly with io.ReadAll for better performance with large inputs
		stdinBytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			logger.Error("Error reading stdin: %v", err)
			color.Red("Error reading stdin: %v", err)
			return
		}
		stdinContent = strings.TrimSpace(string(stdinBytes))
		logger.Debug("Read %d bytes from stdin", len(stdinBytes))
	} else {
		logger.Debug("No piped input detected")
	}

	// Use the prompt from the ShellPrompt field
	var prompt string
	if opts.ShellPrompt != "" {
		prompt = opts.ShellPrompt
		logger.Debug("Using prompt from -s flag: %s", prompt)
	} else {
		// If no prompt was provided, use a default one
		if stdinContent != "" {
			prompt = "Explain the following:"
			logger.Debug("No explicit prompt provided, using default: %s", prompt)
		} else {
			logger.Error("No input provided via stdin or arguments")
			color.Red("Error: No input provided via stdin or arguments")
			return
		}
	}

	// Combine the prompt with the stdin content if both exist
	if stdinContent != "" {
		input = fmt.Sprintf("%s\n\n```\n%s\n```", prompt, stdinContent)
		logger.Debug("Combined prompt with stdin content (%d total chars)", len(input))
	} else {
		input = prompt
		logger.Debug("Using prompt as input (%d chars)", len(input))
	}

	// Print a message indicating that the model is working
	modelName := client.GetModelName()
	fmt.Printf("Using model: %s (temp: %.1f, max tokens: %d)\n",
		modelName, opts.Temperature, opts.MaxTokens)
	logger.Info("Using model: %s (temp: %.1f, max tokens: %d)",
		modelName, opts.Temperature, opts.MaxTokens)

	// Send the message and get the response
	assistantColor := color.New(color.FgHiMagenta)
	assistantPrompt := assistantColor.PrintfFunc()

	// If output format is plain text, show the streaming response
	if opts.OutputFormat == "text" {
		assistantPrompt("%s Response: ", consts.RobotEmoji)
	}

	logger.Debug("Sending message to %s provider", opts.Provider)
	response, elapsed, err := sendMessageAndLogHistory(client, input, opts, logger)
	if err != nil {
		logger.Error("Error during message processing: %v", err)
		color.Red("\nError: %v", err)
		return
	}
	logger.Info("Response received in %.2f seconds (%d chars)", elapsed.Seconds(), len(response))

	// Format the output according to the requested format
	formattedResponse := formatOutput(response, opts.OutputFormat)
	logger.Debug("Formatted response using %s format", opts.OutputFormat)

	// For non-text formats or if the format isn't text, print the formatted output
	if opts.OutputFormat != "text" {
		fmt.Println(formattedResponse)
	} else {
		fmt.Println() // Just add a newline for text format since response was already streamed
	}

	// Display metrics if verbose mode is enabled
	if opts.Verbose {
		logger.Debug("Displaying metrics (verbose mode enabled)")
		printMetrics(input, response, elapsed, opts.Assess)
	} else if opts.Assess {
		logger.Debug("Running prompt assessment")
		assessment.AssessPrompt(input)
	}

	logger.Debug("Shell mode completed successfully")
}

func Chat(opts *ChatOptions) {
	// Initialize logging
	logger, err := setupLogging(opts)
	if err != nil {
		color.Red("Error setting up logging: %v", err)
		return
	}

	logger.Info("Chat CLI started with provider: %s", opts.Provider)

	// If shell mode is enabled, use ShellMode instead of interactive chat
	if opts.Shell {
		logger.Debug("Running in shell mode")
		ShellMode(opts, logger)
		return
	}

	logger.Debug("Running in interactive chat mode")

	client, err := CreateChatClient(opts.Provider, logger)
	if err != nil {
		logger.Error("Failed to create chat client: %v", err)
		color.Red("Error: %v", err)
		return
	}

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	userColor := color.New(color.FgHiGreen)
	userPrompt := userColor.PrintfFunc()

	clearScreen()
	fmt.Printf("Chat started using model: %s (type 'exit' to quit)\n", client.GetModelName())
	fmt.Println("Type 'clear' to clear the screen")
	fmt.Println("For multiline input, type 'paste' and press Enter")
	fmt.Println("Use '--verbose' or '-v' for metrics, '--assess' or '-a' for prompt assessment")

	logger.Info("Interactive chat session started with model: %s", client.GetModelName())

	for {
		text, exit := handleUserInput(scanner, userPrompt)
		if exit {
			logger.Info("User requested exit")
			break
		}
		if text == "" {
			logger.Debug("Empty input, continuing")
			continue
		}

		logger.Debug("Processing user input (%d chars)", len(text))

		response, elapsed, err := sendMessageAndLogHistory(client, text, opts, logger)
		if err != nil {
			logger.Error("Error during message processing: %v", err)
			color.Red("Error: %v", err)
			continue
		}

		logger.Info("Response received in %.2f seconds (%d chars)", elapsed.Seconds(), len(response))

		fmt.Println()

		if opts.Verbose {
			logger.Debug("Displaying metrics (verbose mode enabled)")
			printMetrics(text, response, elapsed, opts.Assess)
		} else if opts.Assess {
			logger.Debug("Running prompt assessment")
			assessment.AssessPrompt(text)
		}
	}

	logger.Info("Chat session ended")
}
