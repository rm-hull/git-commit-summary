package ui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/cockroachdb/errors"
	"github.com/rm-hull/git-commit-summary/internal/interfaces"
	llmprovider "github.com/rm-hull/git-commit-summary/internal/llm_provider"
)

type Client struct {
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Run(
	ctx context.Context,
	llmProvider llmprovider.Provider,
	gitClient interfaces.GitClient,
	systemPrompt string,
	userMessage string,
) error {
	model := initialModel(ctx, llmProvider, gitClient, systemPrompt, userMessage)
	p := tea.NewProgram(model)

	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	m, ok := finalModel.(*Model)
	if !ok {
		return errors.New("failed to cast model to *Model")
	}

	if m.err != nil {
		return m.err
	}

	if m.action == Abort {
		return interfaces.ErrAborted
	}

	if m.action == Commit {
		err = gitClient.Commit(m.commitMessage)
		if err != nil {
			return err
		}
	}

	return nil
}

