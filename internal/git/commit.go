package git

import (
	"os"
	"os/exec"
)

func Commit(message string) ([]byte, error) {
	tmpfile, err := os.CreateTemp("", "gitmsg-*.txt")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = os.Remove(tmpfile.Name()) // clean up
	}()

	if _, err := tmpfile.WriteString(message); err != nil {
		return nil, err
	}
	if err := tmpfile.Close(); err != nil {
		return nil, err
	}

	return exec.Command("git", "commit", "-F", tmpfile.Name()).CombinedOutput()
}
