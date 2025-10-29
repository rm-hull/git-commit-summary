package llmprovider

import (
	"context"
	"fmt"
	"os"

	openai "github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type OpenAiProvider struct {
	client *openai.Client
	model  string
}

func NewOpenAiProvider(ctx context.Context) (Provider, error) {
	baseURL := os.Getenv("OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = "none"
	}
	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-4o"
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
		option.WithBaseURL(baseURL))

	return &OpenAiProvider{
		client: &client,
		model:  model,
	}, nil
}

func (provider *OpenAiProvider) Call(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	result, err := provider.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Temperature: openai.Float(0.1),
		Model: provider.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userPrompt),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	return result.Choices[0].Message.Content, nil
}

func (provider *OpenAiProvider) Model() string {
	return provider.model
}
