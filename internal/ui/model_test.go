package ui

import (
	"context"
	"testing"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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

func (m *MockLLMProvider) Call(ctx context.Context, model string, prompt string) (string, error) {
	args := m.Called(ctx, model, prompt)
	return args.String(0), args.Error(1)
}

func (m *MockLLMProvider) Model() string {
	args := m.Called()
	return args.String(0)
}

// MockGitClient is a mock implementation of interfaces.GitClient
type MockGitClient struct {
	mock.Mock
}

func (m *MockGitClient) IsInWorkTree() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockGitClient) StagedFiles() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockGitClient) Diff() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockGitClient) Commit(message string) error {
	args := m.Called(message)
	return args.Error(0)
}

func TestModel_Update(t *testing.T) {
	ctx := context.Background()
	mockLLM := new(MockLLMProvider)
	mockGit := new(MockGitClient)

	// Common setup for InitialModel
	initialModel := func() *Model {
		// Explicitly use the types to avoid "imported and not used" warnings
		var _ interfaces.GitClient = mockGit
		var _ llmprovider.Provider = mockLLM
		return InitialModel(ctx, mockLLM, mockGit, "system prompt", "user message")
	}

	t.Run("tea.KeyMsg - CtrlC in showSpinner state", func(t *testing.T) {
		m := initialModel()
		m.state = showSpinner // Ensure initial state is showSpinner

		updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

		assert.Equal(t, Abort, updatedModel.(*Model).action)
		assert.NotNil(t, cmd)
		assert.IsType(t, tea.QuitMsg{}, cmd())
	})

	t.Run("tea.KeyMsg - CtrlC in other states", func(t *testing.T) {
		m := initialModel()
		m.state = showCommitView // Set to a state other than showSpinner

		// Mock the sub-model's Update method
		mockCommitView := new(mockTeaModel)
		mockCommitView.On("Update", mock.Anything).Return(mockCommitView, (tea.Cmd)(nil))
		m.commitView = mockCommitView

		updatedModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

		assert.Equal(t, None, updatedModel.(*Model).action) // Action should not be Abort
		assert.Nil(t, cmd)                                  // No tea.Quit command
		mockCommitView.AssertCalled(t, "Update", tea.KeyMsg{Type: tea.KeyCtrlC})
	})

	t.Run("gitCheckMsg - empty (no staged changes)", func(t *testing.T) {
		m := initialModel()
		m.state = showSpinner // Ensure initial state is showSpinner

		updatedModel, cmd := m.Update(gitCheckMsg{})

		assert.NotNil(t, updatedModel.(*Model).err)
		assert.NotNil(t, cmd)
		assert.IsType(t, tea.QuitMsg{}, cmd())
	})

	t.Run("gitCheckMsg - non-empty (staged changes)", func(t *testing.T) {
		m := initialModel()
		m.state = showSpinner // Ensure initial state is showSpinner

		mockGit.On("Diff").Return("mocked diff content", nil).Once()

		updatedModel, cmd := m.Update(gitCheckMsg{"file1.go", "file2.go"})

		assert.Nil(t, updatedModel.(*Model).err)
		assert.NotNil(t, cmd)
		msg := cmd()
		assert.IsType(t, gitDiffMsg(""), msg)
		assert.Equal(t, gitDiffMsg("mocked diff content"), msg)
		mockGit.AssertExpectations(t)
	})

	t.Run("gitDiffMsg", func(t *testing.T) {
		m := initialModel()
		m.state = showSpinner // Ensure initial state is showSpinner
		mockLLM.On("Model").Return("test-model").Once()
		// The command returned by Update will execute llmProvider.Call later.
		// No need to set mockLLM.On("Call") here.

		diffContent := "diff --git a/file.go b/file.go"
		updatedModel, cmd := m.Update(gitDiffMsg(diffContent))

		assert.Equal(t, diffContent, updatedModel.(*Model).diff)
		assert.Contains(t, updatedModel.(*Model).spinnerMessage, "Generating commit summary (using: test-model)")
		assert.IsType(t, tea.Batch(nil), cmd)
		mockLLM.AssertExpectations(t)
	})

	t.Run("llmResultMsg - with user message", func(t *testing.T) {
		m := initialModel()
		m.state = showSpinner // Ensure initial state is showSpinner
		llmResult := "This is a summary from LLM."
		userMsg := "Additional user message."

		m.userMessage = userMsg // Set user message for this test case

		mockCommitView := new(mockTeaModel)
		// Only expect Init to be called by Update. View is called by Model.View()
		mockCommitView.On("Init").Return(textarea.Blink).Once()
		m.commitView = mockCommitView

		updatedModel, cmd := m.Update(llmResultMsg(llmResult))

		assert.Equal(t, showCommitView, updatedModel.(*Model).state)
		// Assert that the commitView is set, but not its content directly from Update
		assert.NotNil(t, updatedModel.(*Model).commitView)
		assert.IsType(t, (tea.Cmd)(nil), cmd)
		// mockCommitView.AssertExpectations(t)
	})

	t.Run("llmResultMsg - without user message", func(t *testing.T) {
		m := initialModel()
		m.state = showSpinner // Ensure initial state is showSpinner
		llmResult := "This is a summary from LLM."
		m.userMessage = "" // Ensure no user message

		mockCommitView := new(mockTeaModel)
		// Only expect Init to be called by Update. View is called by Model.View()
		mockCommitView.On("Init").Return(textarea.Blink).Once()
		m.commitView = mockCommitView

		updatedModel, cmd := m.Update(llmResultMsg(llmResult))

		assert.Equal(t, showCommitView, updatedModel.(*Model).state)
		// Assert that the commitView is set, but not its content directly from Update
		assert.NotNil(t, updatedModel.(*Model).commitView)
		assert.NotNil(t, cmd)
		// mockCommitView.AssertExpectations(t)
	})

	t.Run("commitMsg", func(t *testing.T) {
		m := initialModel()
		m.state = showCommitView // Ensure state is showCommitView

		commitContent := "feat: new feature"
		updatedModel, cmd := m.Update(commitMsg(commitContent))

		assert.Equal(t, Commit, updatedModel.(*Model).action)
		assert.Equal(t, commitContent, updatedModel.(*Model).commitMessage)
		assert.NotNil(t, cmd)
		assert.IsType(t, tea.QuitMsg{}, cmd())
	})

	t.Run("regenerateMsg", func(t *testing.T) {
		m := initialModel()
		m.state = showCommitView // Ensure state is showCommitView

		updatedModel, cmd := m.Update(regenerateMsg{})

		assert.Equal(t, showRegeneratePrompt, updatedModel.(*Model).state)
		assert.NotNil(t, updatedModel.(*Model).promptView)
		assert.NotNil(t, cmd)
		assert.IsType(t, textinput.Blink(), cmd())
	})

	t.Run("userResponseMsg", func(t *testing.T) {
		m := initialModel()
		m.state = showRegeneratePrompt // Ensure state is showRegeneratePrompt
		mockLLM.On("Model").Return("test-model").Once()
		// The command returned by Update will execute llmProvider.Call later.
		// No need to set mockLLM.On("Call") here.

		userResponse := "make it shorter"
		updatedModel, cmd := m.Update(userResponseMsg(userResponse))

		assert.Equal(t, showSpinner, updatedModel.(*Model).state)
		assert.Contains(t, updatedModel.(*Model).spinnerMessage, "Re-generating commit summary (using: test-model)")
		assert.IsType(t, tea.Batch(nil), cmd) // Should return tea.Batch(m.spinner.Tick, m.generateSummary)
		mockLLM.AssertExpectations(t)
	})

	t.Run("cancelRegenPromptMsg", func(t *testing.T) {
		m := initialModel()
		m.state = showRegeneratePrompt // Ensure state is showRegeneratePrompt

		// Mock the sub-model's Init method
		mockCommitView := new(mockTeaModel)
		mockCommitView.On("Init").Return((tea.Cmd)(nil)).Once()
		m.commitView = mockCommitView

		updatedModel, cmd := m.Update(cancelRegenPromptMsg{})

		assert.Equal(t, showCommitView, updatedModel.(*Model).state)
		assert.Nil(t, cmd) // Should return m.commitView.Init() which is mocked to return nil
		mockCommitView.AssertExpectations(t)
	})

	t.Run("errMsg", func(t *testing.T) {
		m := initialModel()
		m.state = showSpinner // Ensure state is showSpinner

		testErr := errors.New("something went wrong")
		updatedModel, cmd := m.Update(errMsg{err: testErr})

		assert.Equal(t, testErr, updatedModel.(*Model).err)
		assert.NotNil(t, cmd)
		assert.IsType(t, tea.QuitMsg{}, cmd())
	})

	t.Run("abortMsg", func(t *testing.T) {
		m := initialModel()
		m.state = showCommitView // Ensure state is showCommitView

		updatedModel, cmd := m.Update(abortMsg{})

		assert.Equal(t, Abort, updatedModel.(*Model).action)
		assert.NotNil(t, cmd)
		assert.IsType(t, tea.QuitMsg{}, cmd())
	})

	t.Run("spinner.Update for showSpinner state", func(t *testing.T) {
		m := initialModel()
		m.state = showSpinner
		// Spinner's Update method is tested by charmbracelet/bubbles,
		// here we just ensure it's called and returns its cmd.
		// We can't easily mock spinner.Model directly, so we'll check the cmd.
		updatedModel, cmd := m.Update(spinner.TickMsg{})
		assert.NotNil(t, updatedModel)
		assert.IsType(t, spinner.TickMsg{}, cmd())
	})

	t.Run("commitView.Update for showCommitView state", func(t *testing.T) {
		m := initialModel()
		m.state = showCommitView
		mockCommitView := new(mockTeaModel)
		mockCommitView.On("Update", mock.Anything).Return(mockCommitView, (tea.Cmd)(nil)).Once()
		m.commitView = mockCommitView

		testMsg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, cmd := m.Update(testMsg)

		assert.NotNil(t, updatedModel)
		assert.Nil(t, cmd) // Mock returns nil cmd
		mockCommitView.AssertCalled(t, "Update", testMsg)
	})

	t.Run("promptView.Update for showRegeneratePrompt state", func(t *testing.T) {
		m := initialModel()
		m.state = showRegeneratePrompt
		mockPromptView := new(mockTeaModel)
		mockPromptView.On("Update", mock.Anything).Return(mockPromptView, (tea.Cmd)(nil)).Once()
		m.promptView = mockPromptView

		testMsg := tea.KeyMsg{Type: tea.KeyEnter}
		updatedModel, cmd := m.Update(testMsg)

		assert.NotNil(t, updatedModel)
		assert.Nil(t, cmd) // Mock returns nil cmd
		mockPromptView.AssertCalled(t, "Update", testMsg)
	})
}

// mockTeaModel is a generic mock for tea.Model interface
type mockTeaModel struct {
	mock.Mock
}

func (m *mockTeaModel) Init() tea.Cmd {
	args := m.Called()
	return args.Get(0).(tea.Cmd)
}

func (m *mockTeaModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	args := m.Called(msg)
	return args.Get(0).(tea.Model), args.Get(1).(tea.Cmd)
}

func (m *mockTeaModel) View() string {
	args := m.Called()
	return args.String(0)
}
