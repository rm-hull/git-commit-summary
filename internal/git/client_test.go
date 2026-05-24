package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_prepareCommitMessage(t *testing.T) {
	client := NewClient(false)

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
