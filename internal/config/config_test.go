package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	t.Run("Defaults", func(t *testing.T) {
		t.Setenv("LLM_PROVIDER", "")
		t.Setenv("GEMINI_MODEL", "")
		t.Setenv("OPENAI_MODEL", "")

		cfg, err := Load()
		assert.NoError(t, err)
		assert.Equal(t, "google", cfg.LLMProvider)
		assert.Equal(t, "gemini-3-flash-preview", cfg.Model)
		assert.NotEmpty(t, cfg.Prompt)
	})

	t.Run("WithEnvironmentVariables", func(t *testing.T) {
		t.Setenv("LLM_PROVIDER", "openai")
		t.Setenv("GEMINI_MODEL", "gemini-pro")
		t.Setenv("OPENAI_MODEL", "gpt-3.5-turbo")

		cfg, err := Load()
		assert.NoError(t, err)
		assert.Equal(t, "openai", cfg.LLMProvider)
		assert.Equal(t, "gpt-3.5-turbo", cfg.Model)
	})
}

func Test_updateProperties(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name            string
		initialContent  string
		props           map[string]string
		expectedContent string
		expectError     bool
	}{
		{
			name: "handle empty quoted strings",
			initialContent: `KEY1=""
KEY2=
`,
			props: map[string]string{"KEY1": "", "KEY2": ""},
			expectedContent: `KEY1=""
KEY2=""
`,
			expectError: false,
		},
		{
			name:           "empty file, add new properties",
			initialContent: "",
			props:          map[string]string{"KEY1": "value1", "KEY2": "value2"},
			expectedContent: `KEY1="value1"
KEY2="value2"
`,
			expectError: false,
		},
		{
			name: "existing file, update properties",
			initialContent: `KEY1="old_value1"
KEY2=old_value2
COMMENT_LINE=some_comment
`,
			props: map[string]string{"KEY1": "new_value1", "KEY2": "new_value2"},
			expectedContent: `# KEY1="old_value1"
KEY1="new_value1"
# KEY2=old_value2
KEY2="new_value2"
COMMENT_LINE=some_comment
`,
			expectError: false,
		},
		{
			name: "doesnt overwrite values if new value is the same",
			initialContent: `KEY1="old_value1"
KEY2=old_value2
COMMENT_LINE=some_comment
`,
			props: map[string]string{"KEY1": "old_value1", "KEY2": "new_value2"},
			expectedContent: `KEY1="old_value1"
# KEY2=old_value2
KEY2="new_value2"
COMMENT_LINE=some_comment
`,
			expectError: false,
		},
		{
			name: "existing file, add new and update existing",
			initialContent: `EXISTING_KEY="existing_value"
# A comment
`,
			props: map[string]string{"EXISTING_KEY": "updated_value", "NEW_KEY": "new_value"},
			expectedContent: `# EXISTING_KEY="existing_value"
EXISTING_KEY="updated_value"
# A comment
NEW_KEY="new_value"
`,
			expectError: false,
		},
		{
			name: "preserve comments and other lines",
			initialContent: `# This is a comment
SOME_VAR=123
ANOTHER_VAR="hello world"

# Another comment block
`,
			props: map[string]string{"SOME_VAR": "456", "NEW_VAR": "new_val"},
			expectedContent: `# This is a comment
# SOME_VAR=123
SOME_VAR="456"
ANOTHER_VAR="hello world"

# Another comment block
NEW_VAR="new_val"
`,
			expectError: false,
		},
		{
			name: "handle mixed quoting styles",
			initialContent: `QUOTED_KEY="quoted_old"
UNQUOTED_KEY=unquoted_old
`,
			props: map[string]string{"QUOTED_KEY": "quoted_new", "UNQUOTED_KEY": "unquoted_new", "NEW_KEY": "new_val"},
			expectedContent: `# QUOTED_KEY="quoted_old"
QUOTED_KEY="quoted_new"
# UNQUOTED_KEY=unquoted_old
UNQUOTED_KEY="unquoted_new"
NEW_KEY="new_val"
`,
			expectError: false,
		},
		{
			name:           "empty props, no changes",
			initialContent: `KEY1="value1"`,
			props:          map[string]string{},
			expectedContent: `KEY1="value1"
`,
			expectError: false,
		},
		{
			name: "properties with special characters",
			initialContent: `KEY1="value with spaces"
KEY2=value_with_underscores
`,
			props: map[string]string{"KEY1": "new value with spaces", "KEY2": "new_value-with-hyphens"},
			expectedContent: `# KEY1="value with spaces"
KEY1="new value with spaces"
# KEY2=value_with_underscores
KEY2="new_value-with-hyphens"
`,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := filepath.Join(tempDir, tt.name, "config.env")

			// Create initial file content if not empty
			if tt.initialContent != "" {
				err := os.MkdirAll(filepath.Dir(configPath), 0o755)
				assert.NoError(t, err)
				err = os.WriteFile(configPath, []byte(tt.initialContent), 0o644)
				assert.NoError(t, err)
			}

			err := updateProperties(configPath, tt.props)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				content, err := os.ReadFile(configPath)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedContent, string(content))
			}
		})
	}
}
