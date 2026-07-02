package git

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/cockroachdb/errors"
	"github.com/creack/pty"
)

type Client struct {
	addAll bool
}

func NewClient(addAll bool) *Client {
	return &Client{
		addAll: addAll,
	}
}

func (c *Client) IsInWorkTree(ctx context.Context) error {
	result, err := exec.CommandContext(ctx,
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

func (c *Client) ModifiedFiles(ctx context.Context) ([]string, error) {
	args := []string{"diff", "--name-only"}
	if c.addAll {
		args = append(args, "HEAD")
	} else {
		args = append(args, "--staged")
	}

	result, err := exec.CommandContext(ctx, "git", args...).CombinedOutput()
	if err != nil {
		return nil, errors.Wrap(err, "listing modified files failed")
	}

	trimmed := strings.TrimSpace(string(result))
	if trimmed == "" {
		return nil, nil
	}
	return strings.Split(trimmed, "\n"), nil
}

func (c *Client) Diff(ctx context.Context, color, exclude bool) (string, error) {
	args := c.diffArgs(color, exclude)

	cmd := exec.CommandContext(ctx, "git", args...)
	if !color {
		result, err := cmd.CombinedOutput()
		if err != nil {
			return "", errors.Wrap(err, "git diff failed")
		}
		return strings.ReplaceAll(string(result), "\t", "    "), nil
	}

	ptmx, err := pty.Start(cmd)
	if err != nil {
		// fallback to plain pipe if pty fails
		return c.Diff(ctx, false, exclude)
	}
	defer func() { _ = ptmx.Close() }()

	var buf bytes.Buffer
	_, copyErr := io.Copy(&buf, ptmx)
	waitErr := cmd.Wait()

	if copyErr != nil && !errors.Is(copyErr, syscall.EIO) {
		return "", errors.Wrap(copyErr, "reading git diff output failed")
	}

	if waitErr != nil {
		return "", errors.Wrap(waitErr, "waiting for git diff command failed")
	}

	output := buf.String()
	output = strings.ReplaceAll(output, "\r", "")
	return strings.ReplaceAll(output, "\t", "    "), nil
}

func (c *Client) diffArgs(color, exclude bool) []string {
	args := []string{
		"--no-pager",
		"diff",
	}
	if color {
		args = append(args, "--color=always")
	}
	args = append(args,
		"--no-ext-diff",
		"--no-textconv",
	)

	if c.addAll {
		args = append(args, "HEAD")
	} else {
		args = append(args, "--staged")
	}

	if exclude {
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
	}
	return args
}

func (c *Client) prepareCommitMessage(message string, skipCI bool) string {
	if skipCI && !strings.Contains(message, "[skip ci]") {
		lines := strings.SplitN(message, "\n", 2)
		subject := strings.TrimRight(lines[0], " \t")
		if len(lines) == 1 {
			return fmt.Sprintf("%s [skip ci]", subject)
		}
		return fmt.Sprintf("%s [skip ci]\n%s", subject, lines[1])
	}
	return message
}

func (c *Client) Commit(ctx context.Context, message string, skipCI, noVerify bool) error {
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
	if noVerify {
		args = append(args, "--no-verify")
	}
	args = append(args, "-F", tmpfile.Name())

	cmd := exec.CommandContext(ctx, "git", args...)

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

func (c *Client) RecentCommits(ctx context.Context, count int) ([]string, error) {
	if count <= 0 {
		return nil, nil
	}
	args := []string{
		"log",
		fmt.Sprintf("--max-count=%d", count),
		"--format=%s",
		"--no-merges",
	}
	result, err := exec.CommandContext(ctx, "git", args...).CombinedOutput()
	if err != nil {
		return nil, errors.Wrap(err, "fetching recent commits failed")
	}
	trimmed := strings.TrimSpace(string(result))
	if trimmed == "" {
		return nil, nil
	}
	return strings.Split(trimmed, "\n"), nil
}
