package app

import (
	"context"
	"testing"

	"github.com/rm-hull/git-commit-summary/internal/interfaces"
	llmprovider "github.com/rm-hull/git-commit-summary/internal/llm_provider"
	"github.com/stretchr/testify/assert"
)

type mockProvider struct {
	modelName string
	callFunc  func(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

func (m *mockProvider) Call(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	return m.callFunc(ctx, systemPrompt, userPrompt)
}

func (m *mockProvider) Model() string {
	return m.modelName
}

type mockGitClient struct {
	IsInWorkTreeFunc func() error
	StagedFilesFunc  func() ([]string, error)
	DiffFunc         func() (string, error)
	CommitFunc       func(message string) error
}

func (m *mockGitClient) IsInWorkTree() error {
	return m.IsInWorkTreeFunc()
}

func (m *mockGitClient) StagedFiles() ([]string, error) {
	return m.StagedFilesFunc()
}

func (m *mockGitClient) Diff() (string, error) {
	return m.DiffFunc()
}

func (m *mockGitClient) Commit(message string) error {
	return m.CommitFunc(message)
}

type mockUIClient struct {
	RunFunc func(llmProvider llmprovider.Provider, gitClient interfaces.GitClient, prompt, userMessage string) error
}

func (m *mockUIClient) Run(llmProvider llmprovider.Provider, gitClient interfaces.GitClient, prompt, userMessage string) error {
	return m.RunFunc(llmProvider, gitClient, prompt, userMessage)
}

func TestNewApp(t *testing.T) {
	provider := &mockProvider{modelName: "test-model"}
	gitClient := &mockGitClient{}
	uiClient := &mockUIClient{}

	app := NewApp(provider, gitClient, uiClient, "test-prompt")

	assert.NotNil(t, app)
	assert.Equal(t, "test-prompt", app.prompt)
	assert.IsType(t, &mockProvider{}, app.llmProvider)
	assert.IsType(t, &mockGitClient{}, app.git)
	assert.IsType(t, &mockUIClient{}, app.ui)
}

func TestAppRun(t *testing.T) {
	ctx := context.Background()

	t.Run("UIClientRunError", func(t *testing.T) {
		mp := &mockProvider{modelName: "test-model"}
		gitClient := &mockGitClient{}
		uiClient := &mockUIClient{
			RunFunc: func(llmProvider llmprovider.Provider, gitClient interfaces.GitClient, prompt, userMessage string) error {
				return assert.AnError
			},
		}
		app := NewApp(mp, gitClient, uiClient, "prompt")
		err := app.Run(ctx, "")
		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("UIClientRunSuccess", func(t *testing.T) {
		mp := &mockProvider{modelName: "test-model"}
		gitClient := &mockGitClient{}
		uiClient := &mockUIClient{
			RunFunc: func(llmProvider llmprovider.Provider, gitClient interfaces.GitClient, prompt, userMessage string) error {
				return nil
			},
		}
		app := NewApp(mp, gitClient, uiClient, "prompt")
		err := app.Run(ctx, "")
		assert.NoError(t, err)
	})
}
