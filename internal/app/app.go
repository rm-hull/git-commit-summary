package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/galactixx/stringwrap"
	"github.com/gookit/color"
	"github.com/rm-hull/git-commit-summary/internal/git"
	llmprovider "github.com/rm-hull/git-commit-summary/internal/llm_provider"
	"github.com/rm-hull/git-commit-summary/internal/ui"
)

type Git interface {
	Diff() (string, error)
	Commit(message string) error
}

type UI interface {
	TextArea(value string) (string, bool, error)
}

// Verify that structs implement interfaces
var _ Git = (*git.Client)(nil)
var _ UI = (*ui.Client)(nil)

type App struct {
	llmProvider llmprovider.Provider
	git         Git
	ui          UI
	prompt      string
}

func NewApp(provider llmprovider.Provider, gitClient Git, uiClient UI, prompt string) *App {
	return &App{
		llmProvider: provider,
		git:         gitClient,
		ui:          uiClient,
		prompt:      prompt,
	}
}

func (app *App) Run(ctx context.Context, userMessage string) error {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = color.Render(" <magenta>Running git diff</>")
	s.Start()
	defer s.Stop()

	diffOutput, err := app.git.Diff()
	if err != nil {
		return err
	}

	if len(diffOutput) == 0 {
		return errors.New("no changes are staged")
	}

	s.Suffix = color.Sprintf(" <blue>Generating commit summary (using: </><fg=blue;op=bold>%s</><blue>)</>", app.llmProvider.Model())
	text := fmt.Sprintf(app.prompt, diffOutput)

	message, err := app.llmProvider.Call(ctx, "", text)
	if err != nil {
		return err
	}

	s.Stop()

	if userMessage != "" {
		message = fmt.Sprintf("%s\n\n%s", userMessage, message)
	}

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
