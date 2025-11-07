package llmprovider

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/rm-hull/git-commit-summary/internal/config"
	"google.golang.org/genai"
)

type GoogleProvider struct {
	client *genai.Client
	model  string
}

func NewGoogleProvider(ctx context.Context, cfg *config.Config) (Provider, error) {
	// The genai library automatically uses the GEMINI_API_KEY environment variable.
	// The config package has already loaded it from the .env file.
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to initialize Google client, is GEMINI_API_KEY set?")
	}

	return &GoogleProvider{
		client: client,
		model:  cfg.Gemini.Model,
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
		return "", errors.Wrap(err, "failed to generate content:")
	}

	return result.Text(), nil
}

func (gp *GoogleProvider) Model() string {
	return gp.model
}
