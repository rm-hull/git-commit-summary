package llmprovider

import (
	"context"
	"fmt"
)

type Provider interface {
	Call(ctx context.Context, systemPrompt, userPrompt string) (string, error)
	Model() string
}

func NewProvider(ctx context.Context, llmProvider string) (Provider, error) {
	switch llmProvider {
	case "google":
		return NewGoogleProvider(ctx)
	case "openai":
		return NewOpenAiProvider(ctx)
	default:
		return nil, fmt.Errorf("unknown LLM provider: %s", llmProvider)
	}
}
