package llmprovider

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/rm-hull/git-commit-summary/internal/config"
)

type Provider interface {
	Call(ctx context.Context, systemPrompt, userPrompt string) (string, error)
	Model() string
}

func NewProvider(ctx context.Context, cfg *config.Config) (Provider, error) {
	switch cfg.LLMProvider {
	case "google":
		return NewGoogleProvider(ctx, cfg)
	case "openai":
		return NewOpenAiProvider(ctx, cfg)
	default:
		return nil, errors.Newf("unknown LLM provider: %s", cfg.LLMProvider)
	}
}
