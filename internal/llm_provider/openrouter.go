package llmprovider

import (
	"context"

	"github.com/rm-hull/git-commit-summary/internal/config"
	"github.com/cockroachdb/errors"
	openrouter "github.com/revrost/go-openrouter"
)

type OpenRouterProvider struct {
	client *openrouter.Client
	model  string
}

func NewOpenRouterProvider(ctx context.Context, cfg *config.Config) (Provider, error) {
	return &OpenRouterProvider{
		client: openrouter.NewClient(cfg.APIKey, openrouter.WithXTitle("git-commit-summary")),
		model:  cfg.Model,
	}, nil
}

func (provider *OpenRouterProvider) Call(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	result, err := provider.client.CreateChatCompletion(ctx, openrouter.ChatCompletionRequest{
		Model: provider.model,
		Messages: []openrouter.ChatCompletionMessage{
			openrouter.SystemMessage(systemPrompt),
			openrouter.UserMessage(userPrompt),
		},
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to generate content")
	}

	return result.Choices[0].Message.Content.Text, nil
}

func (provider *OpenRouterProvider) Model() string {
	return provider.model
}
