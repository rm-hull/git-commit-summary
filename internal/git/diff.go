package git

import (
	"os/exec"
)

func Diff() ([]byte, error) {
	return exec.Command(
		"git",
		"--no-pager",
		"diff",
		"--no-ext-diff",
		"--no-textconv",
		"--staged",
		"--diff-filter=ACMRTUXB",
		"--",                 // separates options from pathspecs
		".",                  // include everything under the repo root
		":(exclude)*-lock.*", // package-lock.json, pnpm-lock.yaml, etc.
		":(exclude)*.lock",   // yarn.lock, poetry.lock, Cargo.lock, etc.
		":(exclude)**/build/**",
		":(exclude)**/dist/**",
		":(exclude)**/target/**",
		":(exclude)**/out/**",
	).CombinedOutput()
}
