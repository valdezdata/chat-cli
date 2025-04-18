package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Entry represents a single history entry
type Entry struct {
	Timestamp    time.Time   `json:"timestamp"`
	Provider     string      `json:"provider"`
	ModelName    string      `json:"model_name"`
	Prompt       string      `json:"prompt"`
	Response     string      `json:"response"`
	InputTokens  int         `json:"input_tokens"`
	OutputTokens int         `json:"output_tokens"`
	TotalTokens  int         `json:"total_tokens"`
	TimeTaken    float64     `json:"time_taken"`
	Assessment   *Assessment `json:"assessment,omitempty"`
}

// Assessment represents the prompt assessment results
type Assessment struct {
	OverallScore   int                       `json:"overall_score"`
	OverallRating  string                    `json:"overall_rating"`
	CriteriaScores map[string]CriteriaResult `json:"criteria_scores"`
}

// CriteriaResult represents the result for a single assessment criterion
type CriteriaResult struct {
	Score       int    `json:"score"`
	Rating      string `json:"rating"`
	Description string `json:"description"`
}

// History holds a collection of entries
type History struct {
	Entries []Entry `json:"entries"`
}

// GetHistoryFilePath returns the path to the history file
func GetHistoryFilePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	// Create .chat-cli directory if it doesn't exist
	chatDir := filepath.Join(homeDir, ".chat-cli")
	if err := os.MkdirAll(chatDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	return filepath.Join(chatDir, "history.json"), nil
}

// LoadHistory loads the history from the file
func LoadHistory() (*History, error) {
	filePath, err := GetHistoryFilePath()
	if err != nil {
		return nil, err
	}

	// If file doesn't exist, return an empty history
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return &History{Entries: []Entry{}}, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read history file: %w", err)
	}

	var history History
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, fmt.Errorf("failed to parse history file: %w", err)
	}

	return &history, nil
}

// SaveHistory saves the history to the file
func SaveHistory(history *History) error {
	filePath, err := GetHistoryFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write history file: %w", err)
	}

	return nil
}

// AddEntry adds a new entry to the history
func AddEntry(entry Entry) error {
	history, err := LoadHistory()
	if err != nil {
		return err
	}

	// Add the new entry
	history.Entries = append(history.Entries, entry)

	// Save the updated history
	return SaveHistory(history)
}

// ShowHistory displays the history entries
func ShowHistory(count int) error {
	history, err := LoadHistory()
	if err != nil {
		return err
	}

	// Limit the number of entries to display
	startIdx := 0
	if count > 0 && count < len(history.Entries) {
		startIdx = len(history.Entries) - count
	}

	fmt.Println("Chat History:")
	fmt.Println("=============")

	for i, entry := range history.Entries[startIdx:] {
		fmt.Printf("#%d - %s (%s)\n", i+1, entry.Timestamp.Format("2006-01-02 15:04:05"), entry.Provider)
		fmt.Printf("Model: %s\n", entry.ModelName)
		fmt.Printf("Prompt: %s\n", truncateString(entry.Prompt, 100))
		fmt.Printf("Response: %s\n", truncateString(entry.Response, 100))
		fmt.Printf("Tokens: %d input, %d output, %d total\n", entry.InputTokens, entry.OutputTokens, entry.TotalTokens)
		fmt.Printf("Time taken: %.2f seconds\n", entry.TimeTaken)
		if entry.Assessment != nil {
			fmt.Printf("Assessment Score: %d%% (%s)\n", entry.Assessment.OverallScore, entry.Assessment.OverallRating)
		}
		fmt.Println("-------------")
	}

	return nil
}

// ClearHistory deletes all history entries
func ClearHistory() error {
	// Create an empty history
	history := &History{Entries: []Entry{}}

	// Save the empty history (this will overwrite the existing file)
	return SaveHistory(history)
}

// Helper function to truncate long strings
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
