package interfaces

import (
	"context"
	"github.com/cockroachdb/errors"
)

var ErrAborted = errors.New("aborted")

// CommitOptions contains options for the Commit method
type CommitOptions struct {
	SkipCI   bool
	NoVerify bool
	Fixes    string
}

type GitClient interface {
	IsInWorkTree(ctx context.Context) error
	ModifiedFiles(ctx context.Context) ([]string, error)
	Diff(ctx context.Context, color, exclude bool) (string, error)
	DiffCompactSummary(ctx context.Context) (string, error)
	RecentCommits(ctx context.Context, count int) ([]string, error)
	Commit(ctx context.Context, message string, opts CommitOptions) error
}
