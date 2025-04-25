package main

import (
	"fmt"
	"os"

	"github.com/valdezdata/chat-cli/internal/cli"
	"github.com/valdezdata/chat-cli/internal/history"
	"github.com/valdezdata/chat-cli/internal/version"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var opts = cli.ChatOptions{
	Verbose:      false,
	Provider:     cli.ProviderOllama,
	Assess:       false,
	Shell:        false,
	ShellPrompt:  "string",
	Temperature:  0.7,
	MaxTokens:    4000,
	OutputFormat: "text",
	LogLevel:     "info",
	LogToFile:    false,
	LogToConsole: false,
}

var rootCmd = &cobra.Command{
	Use:   "chat-cli",
	Short: "A terminal-based chat application for LLMs",
	Long: `Chat CLI is a terminal-based application for interacting with various 
Large Language Model (LLM) providers including Ollama, OpenAI, Together AI, 
Groq, and SambaNova.

It supports both interactive chat and shell mode (for use in pipelines),
with customizable model parameters and output formats.`,

	Run: func(cmd *cobra.Command, args []string) {
		// Set shell mode if shell prompt is provided
		if opts.ShellPrompt != "" {
			opts.Shell = true
		}
		cli.Chat(&opts)
	},
	Example: `  # Interactive chat with Ollama
  chat-cli

  # Use OpenAI's
  chat-cli -p openai

  # Analyze code from stdin
  cat main.py | chat-cli -s "Explain this code"

  # Generate markdown documentation
  cat *.go | chat-cli -s "Create documentation" -f markdown > docs.md

  # Assess and improve your prompts
  chat-cli -a

  # Show performance metrics
  chat-cli -v

  # Debug mode with verbose logging
  chat-cli --log-level debug --log-file`,
}

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "Show chat history",
	Long:  `Display the history of your chat sessions, including prompts, responses, and metrics.`,
	Run: func(cmd *cobra.Command, args []string) {
		count, _ := cmd.Flags().GetInt("count")
		if err := history.ShowHistory(count); err != nil {
			color.Red("Error displaying history: %v", err)
		}
	},
}

var clearHistoryCmd = &cobra.Command{
	Use:   "clear-history",
	Short: "Clear chat history",
	Long:  `Delete all stored chat history.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := history.ClearHistory(); err != nil {
			color.Red("Error clearing history: %v", err)
		} else {
			color.Green("History cleared successfully")
		}
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Chat CLI %s\n", version.GetVersionInfo())
	},
}

func init() {
	// Basic flags
	rootCmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false, "Enable verbose output with metrics")
	rootCmd.Flags().VarP((*cli.ProviderFlag)(&opts.Provider), "provider", "p", "LLM provider to use (ollama, openai, together, groq, samba)")
	rootCmd.Flags().BoolVarP(&opts.Assess, "assess", "a", false, "Assess prompt quality and structure")
	rootCmd.Flags().StringVarP(&opts.ShellPrompt, "shell", "s", "", "Shell mode with specified prompt (read from stdin)")
	rootCmd.Flags().BoolVarP(&opts.LogToConsole, "log", "l", false, "Show logs in console")

	// Model parameter flags
	rootCmd.Flags().Float64VarP(&opts.Temperature, "temperature", "t", 0.7, "Temperature for response generation (0.0-1.0)")
	rootCmd.Flags().IntVarP(&opts.MaxTokens, "max-tokens", "m", 4000, "Maximum number of tokens in response")
	rootCmd.Flags().StringVarP(&opts.OutputFormat, "format", "f", "text", "Output format (text, json, markdown)")

	// Logging flags - Added
	rootCmd.Flags().StringVar(&opts.LogLevel, "log-level", "info", "Log level (debug, info, warn, error)")
	rootCmd.Flags().BoolVar(&opts.LogToFile, "log-file", false, "Write logs to file")
	rootCmd.Flags().BoolVar(&opts.SkipHistory, "no-history", false, "Don't save this interaction to history")

	// Group flags for better organization
	markFlagGroup(rootCmd, "Basic Options", []string{"verbose", "provider", "assess", "shell"})
	markFlagGroup(rootCmd, "Model Parameters", []string{"temperature", "max-tokens", "format"})
	markFlagGroup(rootCmd, "Logging Options", []string{"log-level", "log-file"})

	// Add history command
	historyCmd.Flags().IntP("count", "n", 10, "Number of history entries to show")
	rootCmd.AddCommand(historyCmd)

	// Add clear-history command
	rootCmd.AddCommand(clearHistoryCmd)

	// Add version command
	rootCmd.AddCommand(versionCmd)

	// Add detailed help for each provider
	rootCmd.SetHelpTemplate(customHelpTemplate)
}

// The rest of your main.go file remains the same
func markFlagGroup(cmd *cobra.Command, group string, flagNames []string) {
	for _, name := range flagNames {
		if flag := cmd.Flags().Lookup(name); flag != nil {
			flag.Annotations = map[string][]string{
				"group": {group},
			}
		}
	}
}

var customHelpTemplate = `{{with .Long}}{{. | trimTrailingWhitespaces}}{{end}}

Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}

Environment Variables:
  OLLAMA_URL        URL of your Ollama server (default: http://localhost:11434)
  OLLAMA_MODEL      Model to use with Ollama (options: mistral, llama, deepseek)
  TOGETHER_API_KEY  API key for Together AI
  TOGETHER_MODEL    Model to use with Together (options: llama-70b, deepseek)
  GROQ_API_KEY      API key for Groq
  GROQ_MODEL        Model to use with Groq (options: gemma)
  SAMBA_API_KEY     API key for SambaNova
  SAMBA_MODEL       Model to use with SambaNova (options: llama-70b)
  OPENAI_API_KEY    API key for OpenAI
  OPENAI_MODEL      Model to use with OpenAI (options: gpt-4.1-nano)
  GEMINI_API_KEY    API key for Google Gemini
  GEMINI_MODEL      Model to use with Gemini (options: gemini-pro, gemini-flash, gemini-flash-lite)
`

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
