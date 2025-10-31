package config

import (
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
		assert.Equal(t, "gemini-2.5-flash-preview-09-2025", cfg.Gemini.Model)
		assert.Equal(t, "gpt-4o", cfg.OpenAI.Model)
		assert.NotEmpty(t, cfg.Prompt)
	})

	t.Run("WithEnvironmentVariables", func(t *testing.T) {
		t.Setenv("LLM_PROVIDER", "openai")
		t.Setenv("GEMINI_MODEL", "gemini-pro")
		t.Setenv("OPENAI_MODEL", "gpt-3.5-turbo")

		cfg, err := Load()
		assert.NoError(t, err)
		assert.Equal(t, "openai", cfg.LLMProvider)
		assert.Equal(t, "gemini-pro", cfg.Gemini.Model)
		assert.Equal(t, "gpt-3.5-turbo", cfg.OpenAI.Model)
	})
}
