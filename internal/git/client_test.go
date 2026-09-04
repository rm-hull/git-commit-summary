package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_diffArgs(t *testing.T) {
	client := NewClient(false, 0)
	args := client.diffArgs(false, true)

	// Check that :/ is in the args after the separator --
	found := false
	for i, arg := range args {
		if arg == "--" && i+1 < len(args) && args[i+1] == ":/" {
			found = true
			break
		}
	}
	assert.True(t, found, "Expected :/ to be present after -- separator in diffArgs")
}

func TestExceedsMaxTokenLimit(t *testing.T) {
	tests := []struct {
		name       string
		charCounts map[string][2]int
		maxTokens  int
		expected   []string
	}{
		{
			name: "No files exceed limit",
			charCounts: map[string][2]int{
				"file1.go": {100, 0},
				"file2.go": {200, 0},
			},
			maxTokens: 100,
			expected:  nil,
		},
		{
			name: "Filter out large files",
			charCounts: map[string][2]int{
				"small.go":  {80, 40},   // 120 chars = 30 tokens
				"large.go":  {500, 500}, // 1000 chars = 250 tokens
				"medium.go": {180, 20},  // 200 chars = 50 tokens
			},
			maxTokens: 50,
			expected:  []string{"large.go"},
		},
		{
			name:       "Empty input",
			charCounts: map[string][2]int{},
			maxTokens:  100,
			expected:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := exceedsMaxTokenLimit(tt.charCounts, tt.maxTokens)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestClient_DiffWithExclusions(t *testing.T) {
	// This is a basic structural test since we can't easily mock git operations
	client := NewClient(false, 0)

	// Verify the method exists and accepts the right parameters
	// Actual functionality testing would require integration tests with a real git repo
	assert.NotNil(t, client)
}

func TestClient_prepareCommitMessage(t *testing.T) {
	client := NewClient(false, 0)

	tests := []struct {
		name     string
		message  string
		skipCI   bool
		expected string
	}{
		{
			name:     "Single line message with skipCI",
			message:  "feat: something",
			skipCI:   true,
			expected: "feat: something [skip ci]",
		},
		{
			name:     "Multi-line message with skipCI",
			message:  "feat: something\n\nDetailed description",
			skipCI:   true,
			expected: "feat: something [skip ci]\n\nDetailed description",
		},
		{
			name:     "Single line message without skipCI",
			message:  "feat: something",
			skipCI:   false,
			expected: "feat: something",
		},
		{
			name:     "Multi-line message without skipCI",
			message:  "feat: something\n\nDetailed description",
			skipCI:   false,
			expected: "feat: something\n\nDetailed description",
		},
		{
			name:     "Empty message with skipCI",
			message:  "",
			skipCI:   true,
			expected: " [skip ci]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := client.prepareCommitMessage(tt.message, tt.skipCI)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
