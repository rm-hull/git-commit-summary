package git

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Client struct{}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) IsInWorkTree() error {
	result, err := exec.Command(
		"git",
		"rev-parse",
		"--is-inside-work-tree",
	).CombinedOutput()
	output := strings.Trim(string(result), "\n")

	if err != nil {
		fmt.Println(output)
		return fmt.Errorf("git rev-parse failed: %w", err)
	}

	if output != "true" {
		return errors.New(output)
	}

	return nil
}

func (c *Client) Diff() (string, error) {
	result, err := exec.Command(
		"git",
		"--no-pager",
		"diff",
		"--no-ext-diff",
		"--no-textconv",
		"--staged",
		"--diff-filter=ACMRTUXBD",
		"--",                 // separates options from pathspecs
		".",                  // include everything under the repo root
		":(exclude)*-lock.*", // package-lock.json, pnpm-lock.yaml, etc.
		":(exclude)*.lock",   // yarn.lock, poetry.lock, Cargo.lock, etc.
		":(exclude)**/build/**",
		":(exclude)**/dist/**",
		":(exclude)**/target/**",
		":(exclude)**/out/**",
		":(exclude)go.sum",
	).CombinedOutput()

	if err != nil {
		return "", fmt.Errorf("git diff failed: %w", err)
	}
	return string(result), nil
}

func (c *Client) Commit(message string) error {
	tmpfile, err := os.CreateTemp("", "gitmsg-*.txt")
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(tmpfile.Name()) // clean up
	}()

	if _, err := tmpfile.WriteString(message); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}
	if err := tmpfile.Close(); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	// Set up git commit command
	cmd := exec.Command("git", "commit", "-F", tmpfile.Name())

	// Connect stdout/stderr of git to our program’s stdout/stderr
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin // allow interactive prompts (e.g., GPG signing, editor, etc.)

	// Run the command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	return nil
}
