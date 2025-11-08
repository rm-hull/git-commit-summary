package app

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cockroachdb/errors"
	"github.com/rm-hull/git-commit-summary/internal/git"
	"github.com/rm-hull/git-commit-summary/internal/interfaces"
	llmprovider "github.com/rm-hull/git-commit-summary/internal/llm_provider"
	"github.com/rm-hull/git-commit-summary/internal/ui"
)

// Verify that git.Client implements GitClient.
var _ interfaces.GitClient = (*git.Client)(nil)

type App struct {
	llmProvider llmprovider.Provider
	git         interfaces.GitClient
	prompt      string
}

func NewApp(provider llmprovider.Provider, git interfaces.GitClient, prompt string) *App {
	return &App{
		llmProvider: provider,
		git:         git,
		prompt:      prompt,
	}
}

func (app *App) Run(ctx context.Context, userMessage string) error {
	model := ui.InitialModel(ctx, app.llmProvider, app.git, app.prompt, userMessage)
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	m, ok := finalModel.(*ui.Model)
	if !ok {
		return errors.New("failed to cast model to *ui.Model")
	}

	if m.Err() != nil {
		return m.Err()
	}

	if m.Action() == ui.Abort {
		return interfaces.ErrAborted
	}

	if m.Action() == ui.Commit {
		err = app.git.Commit(m.CommitMessage())
		if err != nil {
			return err
		}
	}

	return nil
}
