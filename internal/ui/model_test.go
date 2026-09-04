package ui

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/rm-hull/git-commit-summary/internal/interfaces"
	llmprovider "github.com/rm-hull/git-commit-summary/internal/llm_provider"
)

// MockLLMProvider is a mock implementation of llmprovider.Provider
type MockLLMProvider struct {
	mock.Mock
}

func (m *MockLLMProvider) Call(ctx context.Context, systemPrompt string, userPrompt string) (string, error) {
	args := m.Called(ctx, systemPrompt, userPrompt)
	return args.String(0), args.Error(1)
}

func (m *MockLLMProvider) Model() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockLLMProvider) Implement() llmprovider.Provider {
	return m
}

// MockGitClient is a mock implementation of interfaces.GitClient
type MockGitClient struct {
	mock.Mock
}

func (m *MockGitClient) IsInWorkTree(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockGitClient) ModifiedFiles(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockGitClient) Diff(ctx context.Context, color, exclude bool) (string, error) {
	args := m.Called(ctx, color, exclude)
	return args.String(0), args.Error(1)
}

func (m *MockGitClient) DiffCompactSummary(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}

func (m *MockGitClient) Commit(ctx context.Context, message string, opts interfaces.CommitOptions) error {
	args := m.Called(ctx, message, opts)
	return args.Error(0)
}

func (m *MockGitClient) RecentCommits(ctx context.Context, count int) ([]string, error) {
	args := m.Called(ctx, count)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockGitClient) Implement() interfaces.GitClient {
	return m
}

func TestModel_Update(t *testing.T) {
	ctx := context.Background()

	// Common setup for InitialModel
	initialModel := func(llm *MockLLMProvider, git *MockGitClient) *Model {
		return InitialModel(ctx, llm, git, "system prompt", "user message", "", false, false, 5)
	}

	t.Run("tea.KeyMsg - CtrlC in showSpinner state", func(t *testing.T) {
		llm, git := new(MockLLMProvider), new(MockGitClient)
		m := initialModel(llm, git)
		m.state = showSpinner

		updatedModel, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: 'c', Mod: tea.ModCtrl}))

		assert.Equal(t, Abort, updatedModel.(*Model).action)
		assert.NotNil(t, cmd)
		assert.IsType(t, tea.QuitMsg{}, cmd())
	})

	t.Run("tea.KeyMsg - CtrlC in other states", func(t *testing.T) {
		llm, git := new(MockLLMProvider), new(MockGitClient)
		m := initialModel(llm, git)
		m.state = showCommitView

		mockCommitView := new(mockTeaModel)
		mockCommitView.On("Update", mock.MatchedBy(func(msg tea.Msg) bool {
			if s, ok := msg.(fmt.Stringer); ok {
				return s.String() == "ctrl+c"
			}
			return false
		})).Return(mockCommitView, (tea.Cmd)(nil))
		m.commitView = mockCommitView

		updatedModel, cmd := m.Update(tea.KeyPressMsg(tea.Key{Code: 'c', Mod: tea.ModCtrl}))

		assert.Equal(t, None, updatedModel.(*Model).action)
		assert.Nil(t, cmd)
		mockCommitView.AssertCalled(t, "Update", tea.KeyPressMsg(tea.Key{Code: 'c', Mod: tea.ModCtrl}))
	})

	t.Run("gitCheckMsg - empty (no staged changes)", func(t *testing.T) {
		llm, git := new(MockLLMProvider), new(MockGitClient)
		m := initialModel(llm, git)
		m.state = showSpinner

		updatedModel, cmd := m.Update(gitCheckMsg{})

		assert.NotNil(t, updatedModel.(*Model).err)
		assert.NotNil(t, cmd)
		assert.IsType(t, tea.QuitMsg{}, cmd())
	})

	t.Run("gitCheckMsg - non-empty (staged changes)", func(t *testing.T) {
		llm, git := new(MockLLMProvider), new(MockGitClient)
		m := initialModel(llm, git)
		m.state = showSpinner

		git.On("Diff", mock.Anything, false, true).Return("mocked diff content", nil).Once()

		updatedModel, cmd := m.Update(gitCheckMsg{"file1.go", "file2.go"})

		assert.Nil(t, updatedModel.(*Model).err)
		assert.NotNil(t, cmd)
		msg := cmd()
		assert.IsType(t, gitDiffMsg(""), msg)
		assert.Equal(t, gitDiffMsg("mocked diff content"), msg)
		git.AssertExpectations(t)
	})

	t.Run("gitDiffMsg with hint", func(t *testing.T) {
		llm, git := new(MockLLMProvider), new(MockGitClient)
		m := initialModel(llm, git)
		m.state = showSpinner
		m.hint = "prioritize auth flow"
		llm.On("Model").Return("test-model").Once()
		git.On("RecentCommits", mock.Anything, 5).Return([]string{}, nil).Once()

		diffContent := "diff --git a/file.go b/file.go"
		llm.On("Call", mock.Anything, mock.Anything, mock.MatchedBy(func(p string) bool {
			return strings.Contains(p, "CONTEXT HINT: prioritize auth flow")
		})).Return("summary", nil).Once()

		updatedModel, cmd := m.Update(gitDiffMsg(diffContent))

		assert.Nil(t, updatedModel.(*Model).err)
		assert.NotNil(t, cmd)
		cmd()
		llm.AssertExpectations(t)
		git.AssertExpectations(t)
	})

	t.Run("llmResultMsg - with user message", func(t *testing.T) {
		llm, git := new(MockLLMProvider), new(MockGitClient)
		m := initialModel(llm, git)
		m.state = showSpinner
		llmResult := "This is a summary from LLM."
		userMsg := "Additional user message."
		m.userMessage = userMsg

		updatedModel, cmd := m.Update(llmResultMsg{content: llmResult, duration: time.Second})

		assert.Equal(t, showCommitView, updatedModel.(*Model).state)
		assert.NotNil(t, updatedModel.(*Model).commitView)
		assert.NotNil(t, cmd)
		assert.IsType(t, textarea.Blink(), cmd())
	})

	t.Run("llmResultMsg - without user message", func(t *testing.T) {
		llm, git := new(MockLLMProvider), new(MockGitClient)
		m := initialModel(llm, git)
		m.state = showSpinner
		llmResult := "This is a summary from LLM."
		m.userMessage = ""

		updatedModel, cmd := m.Update(llmResultMsg{content: llmResult, duration: time.Second})

		assert.Equal(t, showCommitView, updatedModel.(*Model).state)
		assert.NotNil(t, updatedModel.(*Model).commitView)
		assert.NotNil(t, cmd)
		assert.IsType(t, textarea.Blink(), cmd())
	})

	t.Run("commitMsg", func(t *testing.T) {
		llm, git := new(MockLLMProvider), new(MockGitClient)
		m := initialModel(llm, git)
		m.state = showCommitView

		commitContent := "feat: new feature"
		updatedModel, cmd := m.Update(commitMsg(commitContent))

		assert.Equal(t, Commit, updatedModel.(*Model).action)
		assert.Equal(t, commitContent, updatedModel.(*Model).commitMessage)
		assert.NotNil(t, cmd)
		assert.IsType(t, tea.QuitMsg{}, cmd())
	})

	t.Run("regenerateMsg - initializes promptView with current hint", func(t *testing.T) {
		llm, git := new(MockLLMProvider), new(MockGitClient)
		m := initialModel(llm, git)
		m.state = showCommitView
		m.hint = "current hint"

		updatedModel, _ := m.Update(regenerateMsg{})

		assert.Equal(t, showRegeneratePrompt, updatedModel.(*Model).state)
		pv, ok := updatedModel.(*Model).promptView.(*promptViewModel)
		assert.True(t, ok)
		assert.Equal(t, "current hint", pv.textinput.Value())
	})

	t.Run("userResponseMsg - updates m.hint", func(t *testing.T) {
		llm, git := new(MockLLMProvider), new(MockGitClient)
		m := initialModel(llm, git)
		m.state = showRegeneratePrompt
		llm.On("Model").Return("test-model").Once()
		git.On("RecentCommits", mock.Anything, 5).Return([]string{}, nil).Once()
		llm.On("Call", mock.Anything, mock.Anything, mock.Anything).Return("summary", nil).Once()

		userResponse := "new hint"
		updatedModel, _ := m.Update(userResponseMsg(userResponse))

		assert.Equal(t, showSpinner, updatedModel.(*Model).state)
		assert.Equal(t, "new hint", updatedModel.(*Model).hint)
	})

	t.Run("cancelRegenPromptMsg - re-enables help text", func(t *testing.T) {
		llm, git := new(MockLLMProvider), new(MockGitClient)
		m := initialModel(llm, git)
		m.state = showRegeneratePrompt

		cv, err := initialCommitViewModel("test message", 0)
		assert.NoError(t, err)
		cv.helpText = false
		m.commitView = cv

		updatedModel, cmd := m.Update(cancelRegenPromptMsg{})

		assert.Equal(t, showCommitView, updatedModel.(*Model).state)
		assert.True(t, cv.helpText)
		assert.NotNil(t, cmd)
	})

	t.Run("showDiffSummaryViewMsg - loads compact summary", func(t *testing.T) {
		llm, git := new(MockLLMProvider), new(MockGitClient)
		m := initialModel(llm, git)
		m.state = showCommitView
		git.On("DiffCompactSummary", mock.Anything).Return("file.go | 2 +-", nil).Once()

		updatedModel, cmd := m.Update(showDiffSummaryViewMsg{})

		assert.Equal(t, showDiffView, updatedModel.(*Model).state)
		assert.NotNil(t, cmd)
		assert.Equal(t, diffSummaryMsg("file.go | 2 +-"), cmd())
		git.AssertExpectations(t)
	})

	t.Run("errMsg", func(t *testing.T) {
		llm, git := new(MockLLMProvider), new(MockGitClient)
		m := initialModel(llm, git)
		m.state = showSpinner

		testErr := errors.New("something went wrong")
		updatedModel, cmd := m.Update(errMsg{err: testErr})

		assert.Equal(t, testErr, updatedModel.(*Model).err)
		assert.NotNil(t, cmd)
		assert.IsType(t, tea.QuitMsg{}, cmd())
	})

	t.Run("abortMsg", func(t *testing.T) {
		llm, git := new(MockLLMProvider), new(MockGitClient)
		m := initialModel(llm, git)
		m.state = showCommitView

		updatedModel, cmd := m.Update(abortMsg{})

		assert.Equal(t, Abort, updatedModel.(*Model).action)
		assert.NotNil(t, cmd)
		assert.IsType(t, tea.QuitMsg{}, cmd())
	})

	t.Run("spinner.Update for showSpinner state", func(t *testing.T) {
		llm, git := new(MockLLMProvider), new(MockGitClient)
		m := initialModel(llm, git)
		m.state = showSpinner
		_, cmd := m.Update(spinner.TickMsg{})

		assert.NotNil(t, cmd)
		assert.IsType(t, spinner.TickMsg{}, cmd())
	})

	t.Run("commitView.Update for showCommitView state", func(t *testing.T) {
		llm, git := new(MockLLMProvider), new(MockGitClient)
		m := initialModel(llm, git)
		m.state = showCommitView
		mockCommitView := new(mockTeaModel)
		mockCommitView.On("Update", mock.Anything).Return(mockCommitView, (tea.Cmd)(nil)).Once()
		m.commitView = mockCommitView

		testMsg := tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter})
		updatedModel, cmd := m.Update(testMsg)

		assert.NotNil(t, updatedModel)
		assert.Nil(t, cmd)
		mockCommitView.AssertCalled(t, "Update", testMsg)
	})

	t.Run("promptView.Update for showRegeneratePrompt state", func(t *testing.T) {
		llm, git := new(MockLLMProvider), new(MockGitClient)
		m := initialModel(llm, git)
		m.state = showRegeneratePrompt
		mockPromptView := new(mockTeaModel)
		mockPromptView.On("Update", mock.Anything).Return(mockPromptView, (tea.Cmd)(nil)).Once()
		m.promptView = mockPromptView

		testMsg := tea.KeyPressMsg(tea.Key{Code: tea.KeyEnter})
		updatedModel, cmd := m.Update(testMsg)

		assert.NotNil(t, updatedModel)
		assert.Nil(t, cmd)
		mockPromptView.AssertCalled(t, "Update", testMsg)
	})

	t.Run("llmResultMsg - YOLO mode", func(t *testing.T) {
		llm, git := new(MockLLMProvider), new(MockGitClient)
		m := InitialModel(ctx, llm, git, "system prompt", "user message", "", true, false, 5)
		updatedModel, cmd := m.Update(llmResultMsg{content: "commit summary", duration: time.Second})

		assert.Equal(t, Commit, updatedModel.(*Model).action)
		assert.Contains(t, updatedModel.(*Model).commitMessage, "commit summary")
		assert.NotNil(t, cmd)
		assert.IsType(t, tea.QuitMsg{}, cmd())
	})

	t.Run("llmResultMsg - YOLO mode - empty summary", func(t *testing.T) {
		llm, git := new(MockLLMProvider), new(MockGitClient)
		m := InitialModel(ctx, llm, git, "system prompt", "", "", true, false, 5)
		updatedModel, cmd := m.Update(llmResultMsg{content: "", duration: time.Second})

		assert.NotNil(t, updatedModel.(*Model).err)
		assert.Equal(t, "failed to generate a commit summary", updatedModel.(*Model).err.Error())
		assert.NotNil(t, cmd)
		assert.IsType(t, tea.QuitMsg{}, cmd())
	})

	t.Run("gitDiffMsg - includes recent commits in prompt", func(t *testing.T) {
		llm, git := new(MockLLMProvider), new(MockGitClient)
		m := initialModel(llm, git)
		m.state = showSpinner
		m.recentCommitsCount = 3
		llm.On("Model").Return("test-model").Once()

		recentCommits := []string{
			"feat: first recent commit",
			"fix: second recent commit",
			"docs: third recent commit",
		}
		git.On("RecentCommits", mock.Anything, 3).Return(recentCommits, nil).Once()

		llm.On("Call", mock.Anything, mock.MatchedBy(func(s string) bool {
			return strings.Contains(s, "### Recent Commit Style Examples") &&
				strings.Contains(s, "feat: first recent commit") &&
				strings.Contains(s, "fix: second recent commit") &&
				strings.Contains(s, "docs: third recent commit")
		}), mock.Anything).Return("summary", nil).Once()

		updatedModel, cmd := m.Update(gitDiffMsg("mocked diff content"))

		assert.Nil(t, updatedModel.(*Model).err)
		assert.NotNil(t, cmd)
		cmd()
		git.AssertExpectations(t)
		llm.AssertExpectations(t)
	})
}

// mockTeaModel is a generic mock for tea.Model interface
type mockTeaModel struct {
	mock.Mock
}

func (m *mockTeaModel) Init() tea.Cmd {
	args := m.Called()
	if len(args) > 0 {
		if cmd, ok := args.Get(0).(tea.Cmd); ok {
			return cmd
		}
	}
	return nil
}

func (m *mockTeaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	args := m.Called(msg)
	var model tea.Model = m
	var cmd tea.Cmd

	if len(args) > 0 {
		if m, ok := args.Get(0).(tea.Model); ok {
			model = m
		}
	}
	if len(args) > 1 {
		if c, ok := args.Get(1).(tea.Cmd); ok {
			cmd = c
		}
	}
	return model, cmd
}

func (m *mockTeaModel) View() tea.View {
	args := m.Called()
	if len(args) > 0 {
		if v, ok := args.Get(0).(tea.View); ok {
			return v
		}
		if s, ok := args.Get(0).(string); ok {
			return tea.NewView(s)
		}
	}
	return tea.NewView("")
}

func TestTrimTrailingSpaces(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no trailing spaces", "hello\nworld", "hello\nworld"},
		{"trailing spaces on some lines", "hello  \nworld \n ", "hello\nworld\n"},
		{"tabs and carriage returns", "line1\t\r\nline2  \n", "line1\nline2\n"},
		{"empty string", "", ""},
		{"single line with space", "hello ", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, trimTrailingSpaces(tt.input))
		})
	}
}
