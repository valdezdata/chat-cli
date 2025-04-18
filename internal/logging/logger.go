package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

// Log levels
const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// String representations of log levels
var levelNames = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

// Color codes for different log levels
var levelColors = map[LogLevel]string{
	DEBUG: "\033[36m", // Cyan
	INFO:  "\033[32m", // Green
	WARN:  "\033[33m", // Yellow
	ERROR: "\033[31m", // Red
	FATAL: "\033[35m", // Magenta
}

// ColorReset is the ANSI code to reset colors
const ColorReset = "\033[0m"

// Logger handles logging operations
type Logger struct {
	level     LogLevel
	output    io.Writer
	useColors bool
}

// DefaultConfig returns the default logging configuration
type Config struct {
	Enabled      bool
	Level        LogLevel
	Console      bool
	File         bool
	FilePath     string
	UseColors    bool
	ShowTime     bool
	ShowFileLine bool
}

// DefaultConfig returns the default logging configuration
func DefaultConfig() Config {
	return Config{
		Enabled:   true,
		Level:     INFO,
		Console:   true,
		File:      false,
		FilePath:  "",
		UseColors: true,
		ShowTime:  true,
	}
}

// New creates a new Logger with the specified level and output
func New() *Logger {
	return &Logger{
		level:     INFO,
		output:    os.Stdout,
		useColors: true,
	}
}

// Output returns the current output writer
func (l *Logger) Output() io.Writer {
	return l.output
}

// Level returns the current log level
func (l *Logger) Level() LogLevel {
	return l.level
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// SetOutput sets the output writer
func (l *Logger) SetOutput(w io.Writer) {
	l.output = w
}

// SetColors enables or disables colored output
func (l *Logger) SetColors(useColors bool) {
	l.useColors = useColors
}

// log writes a message at the specified level
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	// Skip if level is below threshold
	if level < l.level {
		return
	}

	// Format timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	// Format message
	msg := fmt.Sprintf(format, args...)

	// Add color if enabled
	levelName := levelNames[level]
	if l.useColors {
		levelName = levelColors[level] + levelName + ColorReset
	}

	// Write to output with proper formatting
	// Adding a newline before the log message if it's going to console
	if l.output == os.Stdout {
		fmt.Fprintf(l.output, "\n[%s] [%s] %s\n", timestamp, levelName, msg)
	} else {
		fmt.Fprintf(l.output, "[%s] [%s] %s\n", timestamp, levelName, msg)
	}

	// Exit program if level is FATAL
	if level == FATAL {
		os.Exit(1)
	}
}

// Debug logs a message at DEBUG level
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs a message at INFO level
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a message at WARN level
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs a message at ERROR level
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Fatal logs a message at FATAL level and exits the program
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
}

// Setup configures logging based on the provided config
func Setup(config Config) (*Logger, error) {
	if !config.Enabled {
		// Return a no-op logger if logging is disabled
		logger := New()
		logger.SetOutput(io.Discard)
		return logger, nil
	}

	logger := New()
	logger.SetLevel(config.Level)
	logger.SetColors(config.UseColors)

	// Set up output writer(s)
	var writers []io.Writer

	// Only add console output if explicitly enabled
	if config.Console {
		writers = append(writers, os.Stdout)
	}

	// Add file output if enabled
	if config.File {
		// Set default log file path if not specified
		if config.FilePath == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return nil, err
			}
			config.FilePath = filepath.Join(homeDir, ".chat-cli", "logs", "chat-cli.log")
		}

		// Ensure log directory exists
		logDir := filepath.Dir(config.FilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}

		// Open log file (create if not exists, append if exists)
		file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}

		writers = append(writers, file)
	}

	// If no writers, use discard
	if len(writers) == 0 {
		logger.SetOutput(io.Discard)
	} else if len(writers) == 1 {
		logger.SetOutput(writers[0])
	} else {
		logger.SetOutput(io.MultiWriter(writers...))
	}

	return logger, nil
}
