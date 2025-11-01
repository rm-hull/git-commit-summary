package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/galactixx/stringwrap"
	"github.com/rm-hull/git-commit-summary/internal/git"
	llmprovider "github.com/rm-hull/git-commit-summary/internal/llm_provider"
	"github.com/rm-hull/git-commit-summary/internal/ui"
)

// ErrAborted is returned when the user aborts the commit message editing.
var ErrAborted = errors.New("aborted")

type GitClient interface {
	IsInWorkTree() error
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

	if err := app.git.IsInWorkTree(); err != nil {
		return err
	}

	app.ui.StartSpinner(" <magenta>Running git diff</>")
	defer app.ui.StopSpinner()

	out, err := app.git.Diff()
	if err != nil {
		return err
	}

	if len(out) == 0 {
		return errors.New("no changes are staged")
	}

	app.ui.UpdateSpinner(fmt.Sprintf(" <blue>Generating commit summary (using: </><fg=blue;op=bold>%s</><blue>)</>", app.llmProvider.Model()))
	text := fmt.Sprintf(app.prompt, out)

	message, err := app.llmProvider.Call(ctx, "", text)
	if err != nil {
		return err
	}

	if userMessage != "" {
		message = fmt.Sprintf("%s\n\n%s", userMessage, message)
	}

	// Dont remove this line: it is important to stop the spinner
	// before rendering the text area below
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
		return ErrAborted
	}
}
