package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cockroachdb/errors"
	"github.com/galactixx/stringwrap"
	"github.com/rm-hull/git-commit-summary/internal/interfaces"
	llmprovider "github.com/rm-hull/git-commit-summary/internal/llm_provider"
)

type sessionState int

const (
	showSpinner sessionState = iota
	showCommitView
	showRegeneratePrompt
)

type (
	gitCheckMsg          []string
	gitDiffMsg           string
	llmResultMsg         string
	commitMsg            string
	errMsg               struct{ err error }
	abortMsg             struct{}
	regenerateMsg        struct{}
	cancelRegenPromptMsg struct{}
	userResponseMsg      string
)

type Action int

const (
	None Action = iota
	Abort
	Commit
)

type Model struct {
	ctx           context.Context
	state         sessionState
	llmProvider   llmprovider.Provider
	gitClient     interfaces.GitClient
	systemPrompt  string
	userMessage   string
	diff          string
	spinner       spinner.Model
	spinnerMsg    string
	commitView    tea.Model
	commitMessage string
	promptView    tea.Model
	action        Action
	err           error
}

func initialModel(
	ctx context.Context,
	llmProvider llmprovider.Provider,
	gitClient interfaces.GitClient,
	systemPrompt string,
	userMessage string,
) *Model {
	return &Model{
		ctx:          ctx,
		state:        showSpinner,
		llmProvider:  llmProvider,
		gitClient:    gitClient,
		systemPrompt: systemPrompt,
		userMessage:  userMessage,
		spinner:      spinner.New(spinner.WithSpinner(spinner.MiniDot)),
		spinnerMsg:   Magenta.Render("Running git commands to determine staged changes..."),
		action:       None,
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.checkGitStatus)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			if m.state == showSpinner {
				m.action = Abort
				return m, tea.Quit
			}
		}

	case gitCheckMsg:
		if len(msg) == 0 {
			m.err = errors.New("no changes are staged")
			return m, tea.Quit
		}
		return m, m.getGitDiff

	case gitDiffMsg:
		m.spinnerMsg = fmt.Sprintf("%s%s%s",
			Blue.Render("Generating commit summary (using: "),
			BoldBlue.Render(m.llmProvider.Model()),
			Blue.Render(")"),
		)
		m.diff = string(msg)
		return m, m.generateSummary(m.diff, "")

	case llmResultMsg:
		m.state = showCommitView
		commitMessage := string(msg)
		if m.userMessage != "" {
			// append the user supplied message
			commitMessage = fmt.Sprintf("%s\n\n%s", commitMessage, m.userMessage)
		}

		// Swerve a bug in https://github.com/galactixx/stringwrap/pull/1
		if commitMessage != "" {
			var err error
			commitMessage, _, err = stringwrap.StringWrap(commitMessage, 72, 4, false)
			if err != nil {
				m.err = err
				return m, tea.Quit
			}
		}
		commitMessage = strings.ReplaceAll(commitMessage, "\n\n\n", "\n\n")
		m.commitView = initialCommitViewModel(commitMessage)
		return m, m.commitView.Init()

	case commitMsg:
		m.action = Commit
		m.commitMessage = string(msg)
		return m, tea.Quit

	case regenerateMsg:
		m.state = showRegeneratePrompt
		m.promptView = initialPromptViewModel(
			Magenta.Render("Add an optional instruction to help shape regenerating the commit summary:"),
			"ENTER to confirm, or ESC to cancel.",
		)

		return m, m.promptView.Init()

	case userResponseMsg:
		m.state = showSpinner
		m.spinnerMsg = fmt.Sprintf("%s%s%s",
			Blue.Render("Re-generating commit summary (using: "),
			BoldBlue.Render(m.llmProvider.Model()),
			Blue.Render(")"),
		)
		return m, tea.Batch(m.spinner.Tick, m.generateSummary(m.diff, string(msg)))

	case cancelRegenPromptMsg:
		m.state = showCommitView
		return m, m.commitView.Init()

	case errMsg:
		m.err = msg.err
		return m, tea.Quit

	case abortMsg:
		m.action = Abort
		return m, tea.Quit
	}

	var cmd tea.Cmd
	switch m.state {
	case showSpinner:
		m.spinner, cmd = m.spinner.Update(msg)
	case showCommitView:
		m.commitView, cmd = m.commitView.Update(msg)
	case showRegeneratePrompt:
		m.promptView, cmd = m.promptView.Update(msg)
	}
	return m, cmd
}

func (m *Model) View() string {
	switch m.state {
	case showSpinner:
		return m.spinner.View() + " " + m.spinnerMsg
	case showCommitView:
		return m.commitView.View() + m.helpText()
	case showRegeneratePrompt:
		return m.commitView.View() + m.promptView.View()
	default:
		return ""
	}
}

func (m *Model) helpText() string {
	return fmt.Sprintf("%s:commit  %s:clear  %s:undo  %s:redo  %s:regen  %s:abort",
		BoldYellow.Render("CTRL-X"),
		BoldYellow.Render("CTRL-K"),
		BoldYellow.Render("CTRL-Z"),
		BoldYellow.Render("CTRL-Y"),
		BoldYellow.Render("CTRL-R"),
		BoldYellow.Render("ESC"))
}

func (m *Model) checkGitStatus() tea.Msg {
	time.Sleep(1000 * time.Millisecond) // Add a small delay
	if err := m.gitClient.IsInWorkTree(); err != nil {
		return errMsg{err}
	}
	stagedFiles, err := m.gitClient.StagedFiles()
	if err != nil {
		return errMsg{err}
	}
	return gitCheckMsg(stagedFiles)
}

func (m *Model) getGitDiff() tea.Msg {
	diff, err := m.gitClient.Diff()
	if err != nil {
		return errMsg{err}
	}
	return gitDiffMsg(diff)
}

func (m *Model) generateSummary(diff string, userMessage string) tea.Cmd {
	return func() tea.Msg {
		text := fmt.Sprintf(m.systemPrompt, diff)
		if userMessage != "" {
			text += "\n\n**IMPORTANT:** " + userMessage
		}
		resp, err := m.llmProvider.Call(m.ctx, "", text)
		if err != nil {
			return errMsg{err}
		}
		return llmResultMsg(resp)
	}
}
