package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/galactixx/stringwrap"
	"github.com/gookit/color"
	"github.com/rm-hull/git-commit-summary/internal/git"
	llmprovider "github.com/rm-hull/git-commit-summary/internal/llm_provider"
	"github.com/rm-hull/git-commit-summary/internal/ui"
)

type GitClient interface {
	Diff() (string, error)
	Commit(message string) error
}

// Verify that git.Client implements GitClient.
var _ GitClient = (*git.Client)(nil)

type UIClient interface {
	TextArea(value string) (string, bool, error)
	StartSpinner(message string)
	UpdateSpinner(message string)
	StopSpinner()
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
	app.ui.StartSpinner(" <magenta>Running git diff</>")
	defer app.ui.StopSpinner()

	out, err := app.git.Diff()
	if err != nil {
		return err
	}

	if len(out) == 0 {
		return errors.New("no changes are staged")
	}

	app.ui.UpdateSpinner(color.Sprintf(" <blue>Generating commit summary (using: </><fg=blue;op=bold>%s</><blue>)</>", app.llmProvider.Model()))
	text := fmt.Sprintf(app.prompt, out)

	message, err := app.llmProvider.Call(ctx, "", text)
	if err != nil {
		return err
	}

	if userMessage != "" {
		message = fmt.Sprintf("%s\n\n%s", userMessage, message)
	}

	app.ui.StopSpinner()

	wrapped, _, err := stringwrap.StringWrap(message, 72, 4, false)
	if err != nil {
		return err
	}

	wrapped = strings.ReplaceAll(wrapped, "\n\n\n", "\n\n")
	edited, accepted, err := app.ui.TextArea(wrapped)
	if err != nil {
		return err
	}

	if accepted {
		return app.git.Commit(edited)
	} else {
		color.Println("<fg=red;op=bold>ABORTED!</>")
		return nil // Or a specific error for abortion
	}
}
