# Chat CLI

A terminal-based chat application for interacting with various LLM providers including Ollama, OpenAI, Together AI, Groq, and SambaNova.

## Features

- Interactive chat with LLMs in your terminal
- Support for multiple providers (Ollama, OpenAI, Together, Groq, SambaNova, Gemini)
- Streaming responses with color-coded outputs
- Shell mode for using the CLI in pipelines (similar to Simon Willison's LLM tool)
- Prompt quality assessment
- Chat history logging and retrieval
- Built-in versioning system
- Metrics display for performance evaluation
- Customizable output formats (text, JSON, markdown)
- Control over model parameters (temperature, max tokens)

## Installation

### Option 1: Install directly with Go

```bash
# Install the latest version
go install github.com/valdezdata/chat-cli@latest

# Or install a specific version
go install github.com/valdezdata/chat-cli@v1.0.0
```

### Option 2: Build from source

Clone the repository and build:

```bash
git clone https://github.com/valdezdata/chat-cli.git
cd chat-cli
make build
```

Or install to your GOPATH:

```bash
git clone https://github.com/valdezdata/chat-cli.git
cd chat-cli
make install
```

## Usage

### Interactive Mode

Simply run the command to start an interactive chat session:

```bash
chat-cli
```

### Shell Mode

Use the `-s` or `--shell` flag to enable shell mode, which allows you to pipe content from other commands and provide a prompt:

```bash
# Basic usage with prompt after the -s flag
cat mycode.py | chat-cli -s "Explain this code"

# You can also use other flags
cat mycode.py | chat-cli -s "Explain this code" --provider groq --verbose

# If you don't provide a prompt, a default one will be used
cat mycode.go | chat-cli -s
```

### Provider Selection

You can select the LLM provider using the `--provider` flag:

```bash
chat-cli --provider ollama    # Default, uses local Ollama server
chat-cli --provider together  # Uses Together AI
chat-cli --provider groq      # Uses Groq
chat-cli --provider samba     # Uses SambaNova
chat-cli --provider openai    # Uses OpenAI
chat-cli --provider gemini    # Uses Google Gemini
```

### Model Control Options

```bash
chat-cli --temperature 0.3      # Set the temperature (0.0-1.0), lower is more deterministic
chat-cli --max-tokens 2000      # Limit the maximum tokens in the response
chat-cli --format json          # Output format (text, json, markdown)
```

### Prompt Assessment

Use the `--assess` or `-a` flag to analyze your prompts:

```bash
chat-cli -a                     # Enable prompt assessment in interactive mode
chat-cli -s "Explain code" -a   # Assess the prompt in shell mode

# The assessment will:
# 1. Evaluate your prompt on multiple criteria (clarity, specificity, etc.)
# 2. Provide a score for each criterion
# 3. Suggest specific improvements
```

### History Management

Chat CLI automatically logs all your interactions to `~/.chat-cli/history.json`.

**Important Security Note:** Be mindful that all prompts and responses are saved to the history file in plain text by default. **Avoid entering highly sensitive information** (passwords, private keys, personal data, etc.) into prompts. If you need to handle sensitive data in a session, use the `--no-history` flag to prevent that specific interaction from being saved.

You can view and manage your chat history using the following commands:

```bash
chat-cli history              # Show your last 10 chat interactions
chat-cli history -n 50        # Show your last 50 chat interactions
chat-cli clear-history        # Delete all stored history
chat-cli --no-history         # Start a session that won't be saved to history
```

Each history entry includes:

- Timestamp
- Provider and model used
- Prompt and response
- Token counts
- Time taken
- Assessment scores (if assessment was enabled)

### Version Information

Check the current version of Chat CLI:

```bash
chat-cli version
```

### Other Options

```bash
chat-cli --verbose         # Show metrics like token counts and speed
chat-cli --help            # Show all available options
```

### Advanced Pipeline Examples

```bash
# Generate JSON from code analysis
cat main.go | chat-cli -s "Analyze this code and find bugs" --format json > analysis.json

# Generate documentation in Markdown format
cat *.go | chat-cli -s "Create API documentation" --format markdown --temperature 0.2 > docs.md

# Process large files efficiently
cat large_dataset.csv | chat-cli -s "Summarize this data" --max-tokens 500

# Chain commands together
git diff | chat-cli -s "Explain these changes" | chat-cli -s "Generate a commit message for these changes"
```

## Environment Variables

Different providers require different API keys and settings:

- `OLLAMA_URL` - URL of your Ollama server (default: `http://localhost:11434`)
- `OLLAMA_MODEL` - Model to use with Ollama (options: `mistral`, `llama`, `deepseek`)
- `TOGETHER_API_KEY` - API key for Together AI
- `TOGETHER_MODEL` - Model to use with Together (options: `llama-70b`, `deepseek`)
- `GROQ_API_KEY` - API key for Groq
- `GROQ_MODEL` - Model to use with Groq (options: `gemma`)
- `SAMBA_API_KEY` - API key for SambaNova
- `SAMBA_MODEL` - Model to use with SambaNova (options: `llama-70b`)
- `OPENAI_API_KEY` - API key for OpenAI
- `OPENAI_MODEL` - Model to use with OpenAI (options: `gpt-4.1-nano`)
- `GEMINI_API_KEY` - API key for Google Gemini
- `GEMINI_MODEL` - Model to use with Gemini (options: `gemini-pro`, `gemini-flash`, `gemini-flash-lite`)

## Development

### Project Structure

```
.
├── go.mod
├── go.sum
├── internal
│   ├── assessment     # Prompt quality assessment
│   ├── cli            # Command-line interface
│   ├── consts         # Constant values
│   ├── history        # Chat history management
│   ├── logging        # Logging utilities
│   ├── providers      # LLM provider implementations
│   ├── utils          # Utility functions (security, validation)
│   └── version        # Version information
├── main.go            # Entry point
├── Makefile           # Build automation
├── README.md
└── tests              # Unit/Integration tests
    ├── assess_test.go
    └── cli_test.go
```

### Adding a New Provider

To add a new provider:

1. Create a new file in `internal/providers`
2. Implement the `ChatInterface` defined in `internal/providers/interface.go`
3. Add the provider to the constants and provider creation logic in `internal/cli/chat.go`

## Contributing

Contributions are welcome! If you'd like to contribute, please follow these steps:

1.  **Fork the repository** on GitHub.
2.  **Clone your fork** locally (`git clone git@github.com:YOUR_USERNAME/chat-cli.git`).
3.  **Create a new branch** for your feature or bug fix (`git checkout -b feature/your-feature-name` or `git checkout -b fix/your-bug-fix`).
4.  **Make your changes.** Ensure your code adheres to the project's style and structure.
5.  **Add tests** for any new functionality or bug fixes.
6.  **Run tests** to ensure everything passes (e.g., `go test ./...` or potentially `make test` if you have a test target in your Makefile).
7.  **Commit your changes** with a clear and concise commit message (`git commit -m "feat: Add new provider XYZ"`).
8.  **Push your branch** to your fork (`git push origin feature/your-feature-name`).
9.  **Open a Pull Request** on the original `valdezdata/chat-cli` repository (replace with your actual upstream repo name).
10. **Describe your changes** clearly in the Pull Request description. Explain the problem you're solving or the feature you're adding.

I'll try to review your PR as soon as possible. Thank you for contributing!

## License

[MIT License](LICENSE)
