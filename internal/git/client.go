package git

import (
	"bytes"
	"context"
	_ "embed"
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
	addAll    bool
	maxTokens int
}

func NewClient(addAll bool, maxTokens int) *Client {
	return &Client{
		addAll:    addAll,
		maxTokens: maxTokens,
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

	// If maxTokens is set, identify and exclude files with excessive churn
	if exclude && c.maxTokens > 0 {
		var excludedFiles []string
		charCounts, err := c.StagedCharDiffs(ctx)
		if err != nil {
			return "", errors.Wrap(err, "failed to calculate character diffs for token limiting")
		}
		excludedFiles = exceedsMaxTokenLimit(charCounts, c.maxTokens)
		for _, file := range excludedFiles {
			args = append(args, ":(exclude)"+file)
		}
	}

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

func (c *Client) StagedCharDiffs(ctx context.Context) (map[string][2]int, error) {
	args := []string{"diff"}
	if c.addAll {
		args = append(args, "HEAD")
	} else {
		args = append(args, "--staged")
	}

	result, err := exec.CommandContext(ctx, "git", args...).CombinedOutput()
	if err != nil {
		return nil, errors.Wrap(err, "git diff failed")
	}

	charCounts := make(map[string][2]int) // [added, deleted]
	diffText := string(result)
	lines := strings.Split(diffText, "\n")
	var currentFile string
	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git ") {
			currentFile = strings.TrimSpace(strings.TrimPrefix(line, "diff --git "))
			if parts := strings.SplitN(currentFile, " b/", 2); len(parts) == 2 {
				currentFile = parts[1]
			} else if parts := strings.SplitN(currentFile, " a/", 2); len(parts) == 2 {
				currentFile = parts[1]
			}
		} else if currentFile != "" {
			var count [2]int
			if val, ok := charCounts[currentFile]; ok {
				count = val
			}
			if len(line) > 0 {
				if line[0] == '+' && !strings.HasPrefix(line, "+++") {
					count[0] += len(line[1:])
				} else if line[0] == '-' && !strings.HasPrefix(line, "---") {
					count[1] += len(line[1:])
				}
				charCounts[currentFile] = count
			}
		}
	}
	return charCounts, nil
}

func exceedsMaxTokenLimit(charCounts map[string][2]int, maxTokens int) []string {
	var exceeds []string
	for file, counts := range charCounts {
		added, deleted := counts[0], counts[1]
		totalChars := added + deleted

		// rough heuristic: assume 1 token is approximately 4 characters
		estimatedTokens := (totalChars + 3) / 4
		if estimatedTokens > maxTokens {
			exceeds = append(exceeds, file)
		}
	}
	return exceeds
}

func (c *Client) DiffCompactSummary(ctx context.Context) (string, error) {
	args := c.diffArgs(true, false)
	args = append(args, "--compact-summary")

	result, err := exec.CommandContext(ctx, "git", args...).CombinedOutput()
	if err != nil {
		return "", errors.Wrap(err, "git diff compact summary failed")
	}
	return strings.ReplaceAll(string(result), "\t", "    "), nil
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
			"--", // separates options from pathspecs
			":/", // include everything under the repo root
		)
		args = append(args, exclusions()...)
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
	cmd.Env = append(os.Environ(), "GIT_COMMIT_SUMMARY_IGNORE_HOOK=1")

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
