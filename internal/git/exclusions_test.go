package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProjectGitCommitSummaryIgnore(t *testing.T) {
	tempDir := t.TempDir()
	repoDir := filepath.Join(tempDir, "repo", "subdir")
	assert.NoError(t, os.MkdirAll(repoDir, 0o755))

	ignoreFile := filepath.Join(tempDir, "repo", ".gitcommitsummaryignore")
	assert.NoError(t, os.WriteFile(ignoreFile, []byte("ignored-file.txt\n"), 0o644))

	t.Chdir(repoDir)

	actual := projectGitCommitSummaryIgnore()
	assert.Equal(t, "ignored-file.txt\n", actual)
}

func TestUserHomeGitCommitSummaryIgnore(t *testing.T) {
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)
	t.Setenv("USERPROFILE", tempHome)

	ignoreFile := filepath.Join(tempHome, ".gitcommitsummaryignore")
	assert.NoError(t, os.WriteFile(ignoreFile, []byte("*.secret\n"), 0o644))

	actual := userHomeGitCommitSummaryIgnore()
	assert.Equal(t, "*.secret\n", actual)
}

func TestValidateIgnorePattern_InvalidGlob(t *testing.T) {
	assert.False(t, validateIgnorePattern("[unclosed.glob"))
}

func TestDedupeExclusions(t *testing.T) {
	excludes := []string{
		":(exclude)foo.txt",
		":(exclude)bar.txt",
		":(exclude)foo.txt",
	}

	result := dedupe(excludes)
	assert.Equal(t, []string{
		":(exclude)foo.txt",
		":(exclude)bar.txt",
	}, result)
}
