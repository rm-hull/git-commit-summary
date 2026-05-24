package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/cockroachdb/errors"
)

type Client struct {
	addAll bool
}

func NewClient(addAll bool) *Client {
	return &Client{
		addAll: addAll,
	}
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
		return errors.Wrap(err, "git rev-parse failed")
	}

	if output != "true" {
		return errors.New(output)
	}

	return nil
}

func (c *Client) ModifiedFiles() ([]string, error) {
	args := []string{"diff", "--name-only"}
	if c.addAll {
		args = append(args, "HEAD")
	} else {
		args = append(args, "--staged")
	}

	result, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		return nil, errors.Wrap(err, "listing modified files failed")
	}

	trimmed := strings.TrimSpace(string(result))
	if trimmed == "" {
		return nil, nil
	}
	return strings.Split(trimmed, "\n"), nil
}

func (c *Client) Diff() (string, error) {
	args := []string{
		"--no-pager",
		"diff",
		"--no-ext-diff",
		"--no-textconv",
	}

	if c.addAll {
		args = append(args, "HEAD")
	} else {
		args = append(args, "--staged")
	}

	args = append(args,
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
	)

	result, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		return "", errors.Wrap(err, "git diff failed")
	}
	return string(result), nil
}

func (c *Client) prepareCommitMessage(message string, skipCI bool) string {
	if skipCI {
		lines := strings.SplitN(message, "\n", 2)
		if len(lines) == 1 {
			return fmt.Sprintf("%s [skip ci]", message)
		}
		return fmt.Sprintf("%s [skip ci]\n%s", lines[0], lines[1])
	}
	return message
}

func (c *Client) Commit(message string, skipCI bool) error {
	message = c.prepareCommitMessage(message, skipCI)

	tmpfile, err := os.CreateTemp("", "gitmsg-*.txt")
	if err != nil {
		return err
	}
	defer func() {
		_ = os.Remove(tmpfile.Name()) // clean up
	}()

	if _, err := tmpfile.WriteString(message); err != nil {
		return errors.Wrap(err, "git commit failed")
	}
	if err := tmpfile.Close(); err != nil {
		return errors.Wrap(err, "git commit failed")
	}

	// Set up git commit command
	args := []string{"commit"}
	if c.addAll {
		args = append(args, "-a")
	}
	args = append(args, "-F", tmpfile.Name())

	cmd := exec.Command("git", args...)

	// Connect stdout/stderr of git to our program’s stdout/stderr
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin // allow interactive prompts (e.g., GPG signing, editor, etc.)

	// Run the command
	if err := cmd.Run(); err != nil {
		return errors.Wrap(err, "git commit failed")
	}

	return nil
}
