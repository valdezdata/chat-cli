package tests

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/valdezdata/chat-cli/internal/assessment"
)

func TestAssessPromptClarity(t *testing.T) {
	// Disable color output for testing
	color.NoColor = true

	tests := []struct {
		name           string
		prompt         string
		expectedScore  int
		expectedStatus string
	}{
		{
			name:           "Clear prompt",
			prompt:         "Write a clear sentence.",
			expectedScore:  5,
			expectedStatus: "Good",
		},
		{
			name:           "Short vague prompt",
			prompt:         "hi",
			expectedScore:  2,
			expectedStatus: "Needs Improvement",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create pipe to capture stdout
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create pipe: %v", err)
			}

			// Save original stdout and redirect
			oldStdout := os.Stdout
			os.Stdout = w
			// Redirect color output too
			oldColorOutput := color.Output
			color.Output = w
			defer func() {
				os.Stdout = oldStdout
				color.Output = oldColorOutput
			}()

			// Run the function
			assessment.AssessPrompt(tt.prompt)

			// Close writer to flush output
			w.Close()

			// Capture output
			var buf bytes.Buffer
			_, err = io.Copy(&buf, r)
			if err != nil {
				t.Fatalf("Failed to read pipe: %v", err)
			}
			output := buf.String()

			// Debug logging
			t.Logf("Captured output: %q", output)

			// Check Clarity line
			expected := fmt.Sprintf("- Clarity [%d/5]: %s.", tt.expectedScore, tt.expectedStatus)
			if !strings.Contains(output, expected) {
				t.Errorf("AssessPrompt(%q) output = %q, want substring %q", tt.prompt, output, expected)
			}
		})
	}
}
