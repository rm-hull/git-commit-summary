package git

import (
	"os/exec"
)

func Diff() ([]byte, error) {
	return exec.Command("git", "--no-pager", "diff", "--no-ext-diff", "--no-textconv", "--staged", "--diff-filter=ACMRTUXB").CombinedOutput()
}
