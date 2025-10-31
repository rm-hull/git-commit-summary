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
	TextAreaFunc      func(value string) (string, bool, error)
	StartSpinnerFunc  func(message string)
	UpdateSpinnerFunc func(message string)
	StopSpinnerFunc   func()
}

func (m *mockUIClient) TextArea(value string) (string, bool, error) {
	return m.TextAreaFunc(value)
}

func (m *mockUIClient) StartSpinner(message string) {
	if m.StartSpinnerFunc != nil {
		m.StartSpinnerFunc(message)
	}
}

func (m *mockUIClient) UpdateSpinner(message string) {
	if m.UpdateSpinnerFunc != nil {
		m.UpdateSpinnerFunc(message)
	}
}

func (m *mockUIClient) StopSpinner() {
	if m.StopSpinnerFunc != nil {
		m.StopSpinnerFunc()
	}
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
		mp := &mockProvider{modelName: "test-model"}
		gitClient := &mockGitClient{
			DiffFunc: func() (string, error) { return "", assert.AnError },
		}
		uiClient := &mockUIClient{
			StartSpinnerFunc:  func(message string) {},
			UpdateSpinnerFunc: func(message string) {},
			StopSpinnerFunc:   func() {},
		}
		app := NewApp(mp, gitClient, uiClient, "prompt")
		err := app.Run(ctx, "")
		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("NoStagedChanges", func(t *testing.T) {
		mp := &mockProvider{modelName: "test-model"}
		gitClient := &mockGitClient{
			DiffFunc: func() (string, error) { return "", nil },
		}
		uiClient := &mockUIClient{
			StartSpinnerFunc:  func(message string) {},
			UpdateSpinnerFunc: func(message string) {},
			StopSpinnerFunc:   func() {},
		}
		app := NewApp(mp, gitClient, uiClient, "prompt")
		err := app.Run(ctx, "")
		assert.Error(t, err)
		assert.EqualError(t, err, "no changes are staged")
	})

	t.Run("LLMCallError", func(t *testing.T) {
		mp := &mockProvider{
			modelName: "test-model",
			callFunc:  func(ctx context.Context, systemPrompt, userPrompt string) (string, error) { return "", assert.AnError },
		}
		gitClient := &mockGitClient{
			DiffFunc: func() (string, error) { return "diff output", nil },
		}
		uiClient := &mockUIClient{
			StartSpinnerFunc:  func(message string) {},
			UpdateSpinnerFunc: func(message string) {},
			StopSpinnerFunc:   func() {},
		}
		app := NewApp(mp, gitClient, uiClient, "prompt")
		err := app.Run(ctx, "")
		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("TextAreaError", func(t *testing.T) {
		mp := &mockProvider{
			modelName: "test-model",
			callFunc:  func(ctx context.Context, systemPrompt, userPrompt string) (string, error) { return "llm message", nil },
		}
		gitClient := &mockGitClient{
			DiffFunc: func() (string, error) { return "diff output", nil },
		}
		uiClient := &mockUIClient{
			TextAreaFunc:      func(value string) (string, bool, error) { return "", false, assert.AnError },
			StartSpinnerFunc:  func(message string) {},
			UpdateSpinnerFunc: func(message string) {},
			StopSpinnerFunc:   func() {},
		}
		app := NewApp(mp, gitClient, uiClient, "prompt")
		err := app.Run(ctx, "")
		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("UserAborted", func(t *testing.T) {
		mp := &mockProvider{
			modelName: "test-model",
			callFunc:  func(ctx context.Context, systemPrompt, userPrompt string) (string, error) { return "llm message", nil },
		}
		gitClient := &mockGitClient{
			DiffFunc: func() (string, error) { return "diff output", nil },
		}
		uiClient := &mockUIClient{
			TextAreaFunc:      func(value string) (string, bool, error) { return "", false, nil },
			StartSpinnerFunc:  func(message string) {},
			UpdateSpinnerFunc: func(message string) {},
			StopSpinnerFunc:   func() {},
		}
		app := NewApp(mp, gitClient, uiClient, "prompt")
		err := app.Run(ctx, "")
		assert.NoError(t, err)
	})

	t.Run("CommitError", func(t *testing.T) {
		mp := &mockProvider{
			modelName: "test-model",
			callFunc:  func(ctx context.Context, systemPrompt, userPrompt string) (string, error) { return "llm message", nil },
		}
		gitClient := &mockGitClient{
			DiffFunc:   func() (string, error) { return "diff output", nil },
			CommitFunc: func(message string) error { return assert.AnError },
		}
		uiClient := &mockUIClient{
			TextAreaFunc:      func(value string) (string, bool, error) { return "edited message", true, nil },
			StartSpinnerFunc:  func(message string) {},
			UpdateSpinnerFunc: func(message string) {},
			StopSpinnerFunc:   func() {},
		}
		app := NewApp(mp, gitClient, uiClient, "prompt")
		err := app.Run(ctx, "")
		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)
	})

	t.Run("Success", func(t *testing.T) {
		mp := &mockProvider{
			modelName: "test-model",
			callFunc:  func(ctx context.Context, systemPrompt, userPrompt string) (string, error) { return "llm message", nil },
		}
		gitClient := &mockGitClient{
			DiffFunc:   func() (string, error) { return "diff output", nil },
			CommitFunc: func(message string) error { return nil },
		}
		uiClient := &mockUIClient{
			TextAreaFunc:      func(value string) (string, bool, error) { return "edited message", true, nil },
			StartSpinnerFunc:  func(message string) {},
			UpdateSpinnerFunc: func(message string) {},
			StopSpinnerFunc:   func() {},
		}
		app := NewApp(mp, gitClient, uiClient, "prompt")
		err := app.Run(ctx, "")
		assert.NoError(t, err)
	})
}
