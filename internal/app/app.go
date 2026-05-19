package app

import (
	"context"
	"fmt"

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

func (app *App) Run(ctx context.Context, userMessage string, yolo bool) error {
	model := ui.InitialModel(ctx, app.llmProvider, app.git, app.prompt, userMessage, yolo)
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	m, ok := finalModel.(*ui.Model)
	if !ok {
		return errors.New("failed to cast model to *ui.Model")
	}

	// If a newer version was detected at startup, notify now (after UI completes)
	if latest := m.LatestVersion(); latest != "" {
		fmt.Printf("%s new version (%s) of %s is available\n",
			ui.Blue.Bold(true).Render("NOTICE:"),
			latest,
			ui.WhiteBold.Render("git-commit-summary"))
	}

	if m.Err() != nil {
		return m.Err()
	}

	if m.Action() == ui.Abort {
		return interfaces.ErrAborted
	}

	if m.Action() == ui.Commit {
		if yolo {
			fmt.Println(ui.Green.Bold(true).Render("COMMIT MESSAGE:"))
			fmt.Println(m.CommitMessage())
			fmt.Println()
		}
		err = app.git.Commit(m.CommitMessage())
		if err != nil {
			return err
		}
	}

	return nil
}
