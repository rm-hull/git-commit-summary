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
	"github.com/rm-hull/git-commit-summary/internal/config"
	"github.com/rm-hull/git-commit-summary/internal/git"
	llmprovider "github.com/rm-hull/git-commit-summary/internal/llm_provider"
	"github.com/rm-hull/git-commit-summary/internal/ui"
)

type App struct {
	llmProvider llmprovider.Provider
	prompt      string
}

func NewApp(ctx context.Context, cfg *config.Config) (*App, error) {
	provider, err := llmprovider.NewProvider(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &App{
		llmProvider: provider,
		prompt:      cfg.Prompt,
	}, nil
}

func (a *App) Run(ctx context.Context, userMessage string) error {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = color.Render(" <magenta>Running git diff</>")
	s.Start()
	defer s.Stop()

	out, err := git.Diff()
	if err != nil {
		return err
	}

	if len(out) == 0 {
		return errors.New("no changes are staged")
	}

	s.Suffix = color.Sprintf(" <blue>Generating commit summary (using: </><fg=blue;op=bold>%s</><blue>)</>", a.llmProvider.Model())
	text := fmt.Sprintf(a.prompt, out)

	message, err := a.llmProvider.Call(ctx, "", text)
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
	edited, accepted, err := ui.TextArea(wrapped)
	if err != nil {
		return err
	}

	if accepted {
		return git.Commit(edited)
	} else {
		color.Println("<fg=red;op=bold>ABORTED!</>")
		return nil // Or a specific error for abortion
	}
}
