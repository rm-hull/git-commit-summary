package app

import (
	"context"
	"fmt"

	"os"

	tea "charm.land/bubbletea/v2"
	"github.com/cockroachdb/errors"
	"github.com/rm-hull/git-commit-summary/internal/git"
	"github.com/rm-hull/git-commit-summary/internal/interfaces"
	llmprovider "github.com/rm-hull/git-commit-summary/internal/llm_provider"
	"github.com/rm-hull/git-commit-summary/internal/ui"
)

// Verify that git.Client implements GitClient.
var _ interfaces.GitClient = (*git.Client)(nil)

type App struct {
	llmProvider           llmprovider.Provider
	git                   interfaces.GitClient
	prompt                string
	includeProjectContext bool
	recentCommitsCount    int
}

type RunOptions struct {
	CommitMsgFile string
	UserMessage   string
	Hint          string
	Yolo          bool
	SkipCI        bool
	NoVerify      bool
}

func (ro *RunOptions) HandleError(err error) {
	if err != nil {
		if errors.Is(err, interfaces.ErrAborted) {
			fmt.Println(ui.BoldRed.Render("ABORTED!"))
			exitcode := 0
			if ro.CommitMsgFile != "" {
				exitcode = 1
			}
			os.Exit(exitcode)
		} else {
			prefix := ui.BoldRed.Render("ERROR:")
			fmt.Fprintf(os.Stderr, "%s %v\n", prefix, err)
			os.Exit(1)
		}
	}
}

func NewApp(provider llmprovider.Provider, git interfaces.GitClient, prompt string, includeProjectContext bool, recentCommitsCount int) *App {
	return &App{
		llmProvider:           provider,
		git:                   git,
		prompt:                prompt,
		includeProjectContext: includeProjectContext,
		recentCommitsCount:    recentCommitsCount,
	}
}

func (app *App) Run(ctx context.Context, opts RunOptions) error {
	model := ui.InitialModel(ctx, app.llmProvider, app.git, app.prompt, opts.UserMessage, opts.Hint, opts.Yolo, app.includeProjectContext, app.recentCommitsCount)
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
		fmt.Printf("%s a new version of %s is available (%s)\n",
			ui.Blue.Bold(true).Render("NOTICE:"),
			ui.WhiteBold.Render("git-commit-summary"),
			latest)
	}

	if m.Err() != nil {
		return m.Err()
	}

	if m.Action() == ui.Abort {
		return interfaces.ErrAborted
	}

	if m.Action() == ui.Commit {
		if opts.CommitMsgFile != "" {
			err = os.WriteFile(opts.CommitMsgFile, []byte(m.CommitMessage()), 0644)
			if err != nil {
				return errors.Wrap(err, "failed to write commit message to file")
			}
		} else {
			if opts.Yolo {
				fmt.Println(ui.Green.Bold(true).Render("COMMIT MESSAGE:"))
				fmt.Println(m.CommitMessage())
				fmt.Println()
			}
			err = app.git.Commit(ctx, m.CommitMessage(), opts.SkipCI, opts.NoVerify)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
