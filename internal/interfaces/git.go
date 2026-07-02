package interfaces

import (
	"context"
	"github.com/cockroachdb/errors"
)

var ErrAborted = errors.New("aborted")

type GitClient interface {
	IsInWorkTree(ctx context.Context) error
	ModifiedFiles(ctx context.Context) ([]string, error)
	Diff(ctx context.Context, color, exclude bool) (string, error)
	RecentCommits(ctx context.Context, count int) ([]string, error)
	Commit(ctx context.Context, message string, skipCI, noVerify bool) error
}
