package interfaces

import "github.com/cockroachdb/errors"

var ErrAborted = errors.New("aborted")

type GitClient interface {
	IsInWorkTree() error
	StagedFiles() ([]string, error)
	Diff() (string, error)
	Commit(message string) error
}
