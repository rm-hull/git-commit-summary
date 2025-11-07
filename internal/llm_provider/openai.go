package llmprovider

import (
	"context"

	"github.com/rm-hull/git-commit-summary/internal/config"

	"github.com/cockroachdb/errors"
	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type OpenAiProvider struct {
	client *openai.Client
	model  string
}

func NewOpenAiProvider(ctx context.Context, cfg *config.Config) (Provider, error) {
	client := openai.NewClient(
		option.WithAPIKey(cfg.OpenAI.APIKey),
		option.WithBaseURL(cfg.OpenAI.BaseURL))

	return &OpenAiProvider{
		client: &client,
		model:  cfg.OpenAI.Model,
	}, nil
}

func (provider *OpenAiProvider) Call(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	result, err := provider.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Temperature: openai.Float(0.1),
		Model:       provider.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userPrompt),
		},
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to generate content")
	}

	return result.Choices[0].Message.Content, nil
}

func (provider *OpenAiProvider) Model() string {
	return provider.model
}
