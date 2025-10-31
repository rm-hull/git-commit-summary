package llmprovider

import (
	"context"
	"testing"

	"github.com/rm-hull/git-commit-summary/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewProvider(t *testing.T) {
	t.Run("GoogleProvider", func(t *testing.T) {
		t.Setenv("GEMINI_API_KEY", "dummy-gemini-key")
		cfg := &config.Config{
			LLMProvider: "google",
			Gemini:      config.GeminiConfig{Model: "gemini-test-model"},
		}
		provider, err := NewProvider(context.Background(), cfg)
		assert.NoError(t, err)
		assert.IsType(t, &GoogleProvider{}, provider)
		assert.Equal(t, "gemini-test-model", provider.Model())
	})

	t.Run("OpenAIProvider", func(t *testing.T) {
		t.Setenv("OPENAI_API_KEY", "dummy-openai-key")
		cfg := &config.Config{
			LLMProvider: "openai",
			OpenAI:      config.OpenAIConfig{Model: "openai-test-model"},
		}
		provider, err := NewProvider(context.Background(), cfg)
		assert.NoError(t, err)
		assert.IsType(t, &OpenAiProvider{}, provider)
		assert.Equal(t, "openai-test-model", provider.Model())
	})

	t.Run("UnknownProvider", func(t *testing.T) {
		cfg := &config.Config{LLMProvider: "unknown"}
		_, err := NewProvider(context.Background(), cfg)
		assert.Error(t, err)
		assert.EqualError(t, err, "unknown LLM provider: unknown")
	})
}
