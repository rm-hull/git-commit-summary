package llmprovider

import (
	"context"
	"fmt"
	"os"
)

type Provider interface {
	Call(ctx context.Context, systemPrompt, userPrompt string) (string, error)
	Model() string
}

func NewProvider(ctx context.Context) (Provider, error) {
	provider := os.Getenv("LLM_PROVIDER")
	if provider == "" {
		provider = "google"
	}

	switch provider {
	case "google":
		return NewGoogleProvider(ctx)
	case "openai":
		return NewOpenAiProvider(ctx)
	default:
		return nil, fmt.Errorf("unknown provider: %s", provider)
	}
}
