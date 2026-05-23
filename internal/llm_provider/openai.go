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
		option.WithAPIKey(cfg.APIKey),
		option.WithBaseURL(cfg.BaseURL))

	return &OpenAiProvider{
		client: &client,
		model:  cfg.Model,
	}, nil
}

func (provider *OpenAiProvider) Call(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage(userPrompt),
	}
	if systemPrompt != "" {
		messages = append([]openai.ChatCompletionMessageParamUnion{openai.SystemMessage(systemPrompt)}, messages...)
	}

	result, err := provider.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Temperature: openai.Float(0.1),
		Model:       provider.model,
		Messages:    messages,
	})
	if err != nil {
		return "", errors.Wrap(err, "failed to generate content")
	}

	return result.Choices[0].Message.Content, nil
}

func (provider *OpenAiProvider) Model() string {
	return provider.model
}
