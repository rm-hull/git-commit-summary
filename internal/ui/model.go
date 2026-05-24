package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cockroachdb/errors"
	"github.com/earthboundkid/versioninfo/v2"
	"github.com/galactixx/stringwrap"
	"github.com/rm-hull/git-commit-summary/internal/git"
	"github.com/rm-hull/git-commit-summary/internal/interfaces"
	llmprovider "github.com/rm-hull/git-commit-summary/internal/llm_provider"
	versionpkg "github.com/rm-hull/git-commit-summary/internal/version"
)

type sessionState int

const (
	showSpinner sessionState = iota
	showCommitView
	showRegeneratePrompt
	showDiffView
)

type (
	gitCheckMsg  []string
	gitDiffMsg   string
	llmResultMsg struct {
		content  string
		duration time.Duration
	}
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
	ctx            context.Context
	state          sessionState
	llmProvider    llmprovider.Provider
	gitClient      interfaces.GitClient
	systemPrompt   string
	userMessage    string
	hint           string
	diff           string
	spinner        spinner.Model
	spinnerMessage string
	latestVersion  string
	width, height  int
	commitView     tea.Model
	diffView       tea.Model
	diffLoaded     bool
	commitMessage  string
	promptView     tea.Model
	action         Action
	err            error
	yolo           bool
}

func InitialModel(
	ctx context.Context,
	llmProvider llmprovider.Provider,
	gitClient interfaces.GitClient,
	systemPrompt string,
	userMessage string,
	hint string,
	yolo bool,
) *Model {
	return &Model{
		ctx:            ctx,
		state:          showSpinner,
		llmProvider:    llmProvider,
		gitClient:      gitClient,
		systemPrompt:   systemPrompt,
		userMessage:    userMessage,
		hint:           hint,
		spinner:        spinner.New(spinner.WithSpinner(spinner.MiniDot)),
		spinnerMessage: Magenta.Render("Checking whether a newer version exists..."),
		diffView:       initialDiffViewModel(72+2, 20),
		action:         None,
		yolo:           yolo,
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.checkLatestVersion)
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
		case tea.KeyCtrlD:
			if m.state == showCommitView {
				m.state = showDiffView
				if !m.diffLoaded {
					return m, m.getDiffWithColor
				}
				return m, nil
			} else if m.state == showDiffView {
				m.state = showCommitView
				return m, nil
			}
		case tea.KeyCtrlA:
			if m.state == showCommitView {
				_, cmd := m.commitView.Update(msg)
				return m, cmd
			}
		case tea.KeyCtrlR:
			if m.state == showCommitView {
				_, cmd := m.commitView.Update(msg)
				return m, cmd
			}
		case tea.KeyCtrlP:
			if m.state == showCommitView {
				// This triggers preview in commitView, but we aren't changing state here.
				// However, the user might want to toggle preview while in commit view.
				// The current implementation of commitView handles preview.
				// If we want to keep the same behavior, we just let commitView handle it.
				_, cmd := m.commitView.Update(msg)
				return m, cmd
			}
		case tea.KeyEsc:
			if m.state == showDiffView {
				m.state = showCommitView
				return m, nil
			}
		}

	case diffColorMsg:
		m.diffLoaded = true
		var cmd tea.Cmd
		m.diffView, cmd = m.diffView.Update(msg)
		return m, cmd

	case gitCheckMsg:
		if len(msg) == 0 {
			m.err = errors.New("no changes detected")
			return m, tea.Quit
		}
		return m, m.getGitDiff

	case gitDiffMsg:
		m.spinnerMessage = fmt.Sprintf("%s%s%s",
			Blue.Render("Generating commit summary (using: "),
			Blue.Bold(true).Underline(true).Render(m.llmProvider.Model()),
			Blue.Render(")"),
		)
		m.diff = string(msg)
		return m, m.generateSummary(m.diff, "")

	case llmResultMsg:
		commitMessage := msg.content
		if m.userMessage != "" {
			// append the user supplied message
			commitMessage = fmt.Sprintf("%s\n\n%s", commitMessage, m.userMessage)
		}

		// Swerve a bug in https://github.com/galactixx/stringwrap/pull/1
		if commitMessage != "" {
			commitMessage, _, m.err = stringwrap.StringWrap(commitMessage, 72, 4, false)
			if m.err != nil {
				return m, tea.Quit
			}
		}
		commitMessage = strings.ReplaceAll(commitMessage, "\n\n\n", "\n\n")

		if m.yolo {
			if commitMessage == "" {
				m.err = errors.New("failed to generate a commit summary")
				return m, tea.Quit
			}
			m.action = Commit
			m.commitMessage = commitMessage
			return m, tea.Quit
		}

		m.state = showCommitView
		m.commitView, m.err = initialCommitViewModel(commitMessage, msg.duration)
		if m.err != nil {
			return m, tea.Quit
		}
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
		m.spinnerMessage = fmt.Sprintf("%s%s%s",
			Blue.Render("Re-generating commit summary (using: "),
			Blue.Bold(true).Underline(true).Render(m.llmProvider.Model()),
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

	case latestVersionMsg:
		m.latestVersion = string(msg)
		m.spinnerMessage = Magenta.Render("Running git commands to determine modified files...")
		return m, m.checkGitStatus

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Propagate to commitView so it can adjust its internal textarea/viewport
		var cmd tea.Cmd
		if m.commitView != nil {
			m.commitView, cmd = m.commitView.Update(msg)
		}
		return m, cmd
	}

	var cmd tea.Cmd
	switch m.state {
	case showSpinner:
		m.spinner, cmd = m.spinner.Update(msg)
	case showCommitView:
		m.commitView, cmd = m.commitView.Update(msg)
	case showRegeneratePrompt:
		m.promptView, cmd = m.promptView.Update(msg)
	case showDiffView:
		m.diffView, cmd = m.diffView.Update(msg)
	}
	return m, cmd
}

type latestVersionMsg string

func (m *Model) checkLatestVersion() tea.Msg {
	time.Sleep(500 * time.Millisecond) // Add a small delay
	latest, _ := versionpkg.CheckLatest(versioninfo.Short())
	return latestVersionMsg(latest)
}

func (m *Model) LatestVersion() string {
	return m.latestVersion
}

// LatestVersionMsg is handled above to chain into git checks

func (m *Model) View() string {
	switch m.state {
	case showSpinner:
		return m.spinner.View() + " " + m.spinnerMessage
	case showCommitView:
		if m.commitView == nil {
			return m.spinner.View() + " " + m.spinnerMessage
		}
		return m.commitView.View()
	case showRegeneratePrompt:
		if m.commitView == nil || m.promptView == nil {
			return m.spinner.View() + " " + m.spinnerMessage
		}
		return m.commitView.View() + m.promptView.View()
	case showDiffView:
		return m.diffView.View()
	default:
		return ""
	}
}

func (m *Model) checkGitStatus() tea.Msg {
	time.Sleep(500 * time.Millisecond) // Add a small delay
	if err := m.gitClient.IsInWorkTree(); err != nil {
		return errMsg{err}
	}
	modifiedFiles, err := m.gitClient.ModifiedFiles()
	if err != nil {
		return errMsg{err}
	}
	return gitCheckMsg(modifiedFiles)
}

func (m *Model) getDiffWithColor() tea.Msg {
	diff, err := m.gitClient.(*git.Client).DiffWithColor()
	if err != nil {
		return errMsg{err}
	}
	return diffColorMsg(diff)
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
		var systemInstruction string
		var userPrompt string

		// Split the systemPrompt into instructions and the diff template.
		// The prompt.md format is: [Instructions] \n\n Diff follows: \n\n ```diff %s ``` ...
		parts := strings.SplitN(m.systemPrompt, "Diff follows:", 2)
		if len(parts) < 2 {
			// Fallback if "Diff follows:" is not found in the prompt template.
			systemInstruction = ""
			userPrompt = fmt.Sprintf(m.systemPrompt, diff)
		} else {
			systemInstruction = strings.TrimSpace(parts[0])
			userPrompt = fmt.Sprintf(parts[1], diff)
		}

		if m.hint != "" {
			userPrompt += "\n\nCONTEXT HINT: " + m.hint
		}
		if userMessage != "" {
			userPrompt += "\n\n**IMPORTANT:** " + userMessage
		}

		start := time.Now()
		resp, err := m.llmProvider.Call(m.ctx, systemInstruction, userPrompt)
		duration := time.Since(start)
		if err != nil {
			return errMsg{err}
		}
		return llmResultMsg{content: resp, duration: duration}
	}
}

func (m *Model) Err() error {
	return m.err
}

func (m *Model) Action() Action {
	return m.action
}

func (m *Model) CommitMessage() string {
	return m.commitMessage
}
