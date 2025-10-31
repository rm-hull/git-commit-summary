package app

import (
	"context"
	"testing"

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
	DiffFunc   func() (string, error)
	CommitFunc func(message string) error
}

func (m *mockGitClient) Diff() (string, error) {
	return m.DiffFunc()
}

func (m *mockGitClient) Commit(message string) error {
	return m.CommitFunc(message)
}

type mockUIClient struct {
	TextAreaFunc func(value string) (string, bool, error)
}

func (m *mockUIClient) TextArea(value string) (string, bool, error) {
	return m.TextAreaFunc(value)
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

	t.Run("DiffError", func(t *testing.T) {
		gitClient := &mockGitClient{
			DiffFunc: func() (string, error) { return "", assert.AnError },
		}
		app := NewApp(&mockProvider{}, gitClient, &mockUIClient{}, "prompt")
		err := app.Run(ctx, "")
		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("NoStagedChanges", func(t *testing.T) {
		gitClient := &mockGitClient{
			DiffFunc: func() (string, error) { return "", nil },
		}
		app := NewApp(&mockProvider{}, gitClient, &mockUIClient{}, "prompt")
		err := app.Run(ctx, "")
		assert.Error(t, err)
		assert.EqualError(t, err, "no changes are staged")
	})

	t.Run("LLMCallError", func(t *testing.T) {
		gitClient := &mockGitClient{
			DiffFunc: func() (string, error) { return "diff output", nil },
		}
		llmProvider := &mockProvider{
			modelName: "test-model",
			callFunc:  func(ctx context.Context, systemPrompt, userPrompt string) (string, error) { return "", assert.AnError },
		}
		app := NewApp(llmProvider, gitClient, &mockUIClient{}, "prompt")
		err := app.Run(ctx, "")
		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("TextAreaError", func(t *testing.T) {
		gitClient := &mockGitClient{
			DiffFunc: func() (string, error) { return "diff output", nil },
		}
		llmProvider := &mockProvider{
			modelName: "test-model",
			callFunc:  func(ctx context.Context, systemPrompt, userPrompt string) (string, error) { return "llm message", nil },
		}
		uiClient := &mockUIClient{
			TextAreaFunc: func(value string) (string, bool, error) { return "", false, assert.AnError },
		}
		app := NewApp(llmProvider, gitClient, uiClient, "prompt")
		err := app.Run(ctx, "")
		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("UserAborted", func(t *testing.T) {
		gitClient := &mockGitClient{
			DiffFunc: func() (string, error) { return "diff output", nil },
		}
		llmProvider := &mockProvider{
			modelName: "test-model",
			callFunc:  func(ctx context.Context, systemPrompt, userPrompt string) (string, error) { return "llm message", nil },
		}
		uiClient := &mockUIClient{
			TextAreaFunc: func(value string) (string, bool, error) { return "", false, nil },
		}
		app := NewApp(llmProvider, gitClient, uiClient, "prompt")
		err := app.Run(ctx, "")
		assert.NoError(t, err)
	})

	t.Run("CommitError", func(t *testing.T) {
		gitClient := &mockGitClient{
			DiffFunc:   func() (string, error) { return "diff output", nil },
			CommitFunc: func(message string) error { return assert.AnError },
		}
		llmProvider := &mockProvider{
			modelName: "test-model",
			callFunc:  func(ctx context.Context, systemPrompt, userPrompt string) (string, error) { return "llm message", nil },
		}
		uiClient := &mockUIClient{
			TextAreaFunc: func(value string) (string, bool, error) { return "edited message", true, nil },
		}
		app := NewApp(llmProvider, gitClient, uiClient, "prompt")
		err := app.Run(ctx, "")
		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("Success", func(t *testing.T) {
		gitClient := &mockGitClient{
			DiffFunc:   func() (string, error) { return "diff output", nil },
			CommitFunc: func(message string) error { return nil },
		}
		llmProvider := &mockProvider{
			modelName: "test-model",
			callFunc:  func(ctx context.Context, systemPrompt, userPrompt string) (string, error) { return "llm message", nil },
		}
		uiClient := &mockUIClient{
			TextAreaFunc: func(value string) (string, bool, error) { return "edited message", true, nil },
		}
		app := NewApp(llmProvider, gitClient, uiClient, "prompt")
		err := app.Run(ctx, "")
		assert.NoError(t, err)
	})
}
