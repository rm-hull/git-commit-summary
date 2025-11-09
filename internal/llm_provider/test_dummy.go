package llmprovider

import (
	"context"
	_ "embed"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/rm-hull/git-commit-summary/internal/config"
)

//go:embed test/example_commit_message.md
var stockResponse string

type TestDummyProvider struct{}

func NewTestDummy(ctx context.Context, cfg *config.Config) (*TestDummyProvider, error) {
	return &TestDummyProvider{}, nil
}

func (provider *TestDummyProvider) Call(ctx context.Context, systemPrompt, userPrompt string) (string, error) {

	if strings.Contains(userPrompt, "throw error") {
		return "", errors.Newf("simulating a model call failure")
	}

	time.Sleep(1 * time.Second)
	return stockResponse, nil
}

func (provider *TestDummyProvider) Model() string {
	return "test-model"
}
