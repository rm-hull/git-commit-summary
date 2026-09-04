package ui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"github.com/cockroachdb/errors"
	"github.com/earthboundkid/versioninfo/v2"
	"github.com/galactixx/stringwrap"
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
	gitCheckMsg []string
	gitDiffMsg  struct {
		diff              string
		compactSummary    string
		exceededMaxTokens bool
	}
	llmResultMsg struct {
		content  string
		duration time.Duration
	}
	commitMsg              string
	diffColorMsg           string
	diffSummaryMsg         string
	errMsg                 struct{ err error }
	abortMsg               struct{}
	regenerateMsg          struct{}
	cancelRegenPromptMsg   struct{}
	showDiffViewMsg        struct{}
	showDiffSummaryViewMsg struct{}
	cancelDiffViewMsg      struct{}
	userResponseMsg        string
)

type Action int

const (
	None Action = iota
	Abort
	Commit
)

type Model struct {
	ctx                   context.Context
	state                 sessionState
	llmProvider           llmprovider.Provider
	gitClient             interfaces.GitClient
	systemPrompt          string
	userMessage           string
	hint                  string
	diff                  string
	compactSummary        string
	spinner               spinner.Model
	spinnerMessage        string
	latestVersion         string
	commitView            tea.Model
	diffView              tea.Model
	diffLoaded            bool
	commitMessage         string
	promptView            tea.Model
	action                Action
	err                   error
	yolo                  bool
	includeProjectContext bool
	recentCommitsCount    int
	exceededMaxTokens     bool
}

func InitialModel(
	ctx context.Context,
	llmProvider llmprovider.Provider,
	gitClient interfaces.GitClient,
	systemPrompt string,
	userMessage string,
	hint string,
	yolo bool,
	includeProjectContext bool,
	recentCommitsCount int,
) *Model {
	return &Model{
		ctx:                   ctx,
		state:                 showSpinner,
		llmProvider:           llmProvider,
		gitClient:             gitClient,
		systemPrompt:          systemPrompt,
		userMessage:           userMessage,
		hint:                  hint,
		spinner:               spinner.New(spinner.WithSpinner(spinner.MiniDot)),
		spinnerMessage:        Magenta.Render("Checking whether a newer version exists..."),
		diffView:              initialDiffViewModel(72, 20),
		action:                None,
		yolo:                  yolo,
		includeProjectContext: includeProjectContext,
		recentCommitsCount:    recentCommitsCount,
		exceededMaxTokens:     false,
	}
}

func (m *Model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.checkLatestVersion)
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if msg.String() == "ctrl+c" && m.state == showSpinner {
			m.action = Abort
			return m, tea.Quit
		}

	case gitCheckMsg:
		if len(msg) == 0 {
			m.err = errors.New("no changes detected")
			return m, tea.Quit
		}
		return m, m.getGitDiffForLLM

	case gitDiffMsg:
		m.spinnerMessage = fmt.Sprintf("%s%s%s",
			Blue.Render("Generating commit summary (using: "),
			Blue.Bold(true).Underline(true).Render(m.llmProvider.Model()),
			Blue.Render(")"),
		)
		m.diff = msg.diff
		m.compactSummary = msg.compactSummary
		m.exceededMaxTokens = msg.exceededMaxTokens
		return m, m.generateSummary(m.diff, m.compactSummary)

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
		commitMessage = trimTrailingSpaces(commitMessage)
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
		m.commitView, m.err = initialCommitViewModel(commitMessage, msg.duration, m.exceededMaxTokens)
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
			m.hint,
		)

		return m, m.promptView.Init()

	case userResponseMsg:
		m.state = showSpinner
		m.spinnerMessage = fmt.Sprintf("%s%s%s",
			Blue.Render("Re-generating commit summary (using: "),
			Blue.Bold(true).Underline(true).Render(m.llmProvider.Model()),
			Blue.Render(")"),
		)
		m.hint = string(msg)
		return m, tea.Batch(m.spinner.Tick, m.generateSummary(m.diff, m.compactSummary))

	case cancelRegenPromptMsg, cancelDiffViewMsg:
		m.state = showCommitView
		return m, m.commitView.Init()

	case showDiffViewMsg:
		if m.state == showDiffView && m.diffLoaded {
			m.diffView, cmd = m.diffView.Update(msg)
			return m, cmd
		}
		m.state = showDiffView
		if !m.diffLoaded {
			return m, m.getFullDiffWithColor
		}
		return m, m.diffView.Init()

	case showDiffSummaryViewMsg:
		m.state = showDiffView
		return m, m.getDiffCompactSummary

	case diffColorMsg:
		m.diffLoaded = true
		m.diffView, cmd = m.diffView.Update(msg)
		return m, cmd

	case diffSummaryMsg:
		m.diffView, cmd = m.diffView.Update(msg)
		return m, cmd

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
	}

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

func (m *Model) View() tea.View {
	switch m.state {
	case showSpinner:
		return tea.NewView(m.spinner.View() + " " + m.spinnerMessage)
	case showCommitView:
		if m.commitView == nil {
			return tea.NewView(m.spinner.View() + " " + m.spinnerMessage)
		}
		return m.commitView.View()
	case showRegeneratePrompt:
		if m.commitView == nil || m.promptView == nil {
			return tea.NewView(m.spinner.View() + " " + m.spinnerMessage)
		}
		return tea.NewView(m.commitView.View().Content + m.promptView.View().Content)
	case showDiffView:
		return m.diffView.View()
	default:
		return tea.NewView("")
	}
}

func (m *Model) checkGitStatus() tea.Msg {
	time.Sleep(500 * time.Millisecond) // Add a small delay
	ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
	defer cancel()
	if err := m.gitClient.IsInWorkTree(ctx); err != nil {
		return errMsg{err}
	}
	modifiedFiles, err := m.gitClient.ModifiedFiles(ctx)
	if err != nil {
		return errMsg{err}
	}
	return gitCheckMsg(modifiedFiles)
}

func (m *Model) getFullDiffWithColor() tea.Msg {
	ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
	defer cancel()
	diff, _, err := m.gitClient.Diff(ctx, true, false)
	if err != nil {
		return errMsg{err}
	}
	return diffColorMsg(diff)
}

func (m *Model) getDiffCompactSummary() tea.Msg {
	ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
	defer cancel()
	diff, err := m.gitClient.DiffCompactSummary(ctx, true)
	if err != nil {
		return errMsg{err}
	}
	return diffSummaryMsg(diff)
}

func (m *Model) getGitDiffForLLM() tea.Msg {
	ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
	defer cancel()
	diff, exceeded, err := m.gitClient.Diff(ctx, false, true)
	if err != nil {
		return errMsg{err}
	}

	compactSummary, err := m.gitClient.DiffCompactSummary(ctx, false)
	if err != nil {
		return errMsg{err}
	}
	return gitDiffMsg{
		diff:              diff,
		compactSummary:    compactSummary,
		exceededMaxTokens: exceeded,
	}
}

func (m *Model) generateSummary(diff, compactSummary string) tea.Cmd {
	var systemInstruction string
	var userPrompt string

	// Split the systemPrompt into instructions and the diff template.
	parts := strings.SplitN(m.systemPrompt, "### Diff", 2)
	if len(parts) < 2 {
		// Fallback if "### Diff" is not found in the prompt template.
		systemInstruction = ""
		userPrompt = fmt.Sprintf(m.systemPrompt, diff)
	} else {
		systemInstruction = strings.TrimSpace(parts[0])
		userPrompt = fmt.Sprintf(parts[1], diff, compactSummary)
	}

	if m.includeProjectContext {
		if projCtx := m.getProjectContext(); projCtx != "" {
			systemInstruction += "\n\n### Project Context\n````markdown\n" + projCtx + "\n````"
		}
	}

	if m.hint != "" {
		userPrompt += "\n\nCONTEXT HINT: " + m.hint
	}

	if m.recentCommitsCount > 0 {
		ctx, cancel := context.WithTimeout(m.ctx, 30*time.Second)
		defer cancel()
		recent, err := m.gitClient.RecentCommits(ctx, m.recentCommitsCount)
		if err == nil && len(recent) > 0 {
			systemInstruction += "\n\n### Recent Commit Style Examples\n````text\n" + strings.Join(recent, "\n") + "\n````"
		}
	}
	return func() tea.Msg {
		start := time.Now()
		ctx, cancel := context.WithTimeout(m.ctx, 60*time.Second)
		defer cancel()
		resp, err := m.llmProvider.Call(ctx, systemInstruction, userPrompt)
		duration := time.Since(start)
		if err != nil {
			return errMsg{err}
		}
		return llmResultMsg{content: resp, duration: duration}
	}
}

func (m *Model) getProjectContext() string {
	files := []string{".project-context.md", "README.md"}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err == nil {
			// Truncate to 4000 characters to avoid token overflow
			runes := []rune(string(data))
			if len(runes) > 4000 {
				return string(runes[:4000]) + "\n\n[... content truncated ...]"
			}
			return string(runes)
		}
	}
	return ""
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

func trimTrailingSpaces(s string) string {
	lines := strings.Split(s, "\n")
	for i := range lines {
		lines[i] = strings.TrimRight(lines[i], " \t\r")
	}
	return strings.Join(lines, "\n")
}
