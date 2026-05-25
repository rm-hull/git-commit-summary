package interfaces

import "github.com/cockroachdb/errors"

var ErrAborted = errors.New("aborted")

type GitClient interface {
	IsInWorkTree() error
	ModifiedFiles() ([]string, error)
	Diff(color, exclude bool) (string, error)
	Commit(message string, skipCI bool) error
}
