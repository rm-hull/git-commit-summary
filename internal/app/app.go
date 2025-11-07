package app

import (
	"context"

	"github.com/rm-hull/git-commit-summary/internal/git"
	"github.com/rm-hull/git-commit-summary/internal/interfaces"
	"github.com/rm-hull/git-commit-summary/internal/llm_provider"
	"github.com/rm-hull/git-commit-summary/internal/ui"
)

type GitClient = interfaces.GitClient

// Verify that git.Client implements GitClient.
var _ GitClient = (*git.Client)(nil)

type UIClient interface {
	Run(llmProvider llmprovider.Provider, gitClient GitClient, prompt, userMessage string) error
}

// Verify that ui.Client implements UIClient.
var _ UIClient = (*ui.Client)(nil)

type App struct {
	llmProvider llmprovider.Provider
	git         GitClient
	ui          UIClient
	prompt      string
}

func NewApp(provider llmprovider.Provider, git GitClient, ui UIClient, prompt string) *App {
	return &App{
		llmProvider: provider,
		git:         git,
		ui:          ui,
		prompt:      prompt,
	}
}

func (app *App) Run(ctx context.Context, userMessage string) error {
	return app.ui.Run(app.llmProvider, app.git, app.prompt, userMessage)
}
