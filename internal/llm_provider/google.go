package llmprovider

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/genai"
)

type GoogleProvider struct {
	client *genai.Client
	model  string
}

func NewGoogleProvider(ctx context.Context) (*GoogleProvider, error) {
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Google client")
	}
	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-2.5-flash-preview-09-2025"
	}

	return &GoogleProvider{
		client: client,
		model:  model,
	}, nil
}

func (gp *GoogleProvider) Call(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	result, err := gp.client.Models.GenerateContent(
		ctx,
		gp.model,
		genai.Text(userPrompt),
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to generate content: %w", err)
	}

	return result.Text(), nil
}

func (gp *GoogleProvider) Model() string {
	return gp.model
}
