package git

import (
	_ "embed"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
)

//go:embed .gitcommitsummaryignore
var defaultGitCommitSummaryIgnore string

func exclusions() []string {
	var excludes []string
	for _, content := range []string{
		defaultGitCommitSummaryIgnore,
		userHomeGitCommitSummaryIgnore(),
		projectGitCommitSummaryIgnore(),
		xdgConfigGitCommitSummaryIgnore(),
	} {
		for line := range strings.SplitSeq(content, "\n") {
			// remove any comment (hash or double-slash) at the end of the line
			if idx := strings.Index(line, "#"); idx != -1 {
				line = line[:idx]
			}

			if idx := strings.Index(line, "//"); idx != -1 {
				line = line[:idx]
			}

			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			if !validateIgnorePattern(line) {
				continue
			}

			excludes = append(excludes, ":(exclude)"+line)
		}
	}
	return dedupe(excludes)
}

func dedupe(excludes []string) []string {
	seen := make(map[string]struct{})
	var result []string
	for _, exclude := range excludes {
		if _, ok := seen[exclude]; !ok {
			seen[exclude] = struct{}{}
			result = append(result, exclude)
		}
	}
	return result
}

func projectGitCommitSummaryIgnore() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		ignorePath := filepath.Join(cwd, ".gitcommitsummaryignore")
		if data, err := os.ReadFile(ignorePath); err == nil {
			return string(data)
		}

		parent := filepath.Dir(cwd)
		if parent == cwd {
			break
		}
		cwd = parent
	}

	return ""
}

func userHomeGitCommitSummaryIgnore() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return readFileString(filepath.Join(homeDir, ".gitcommitsummaryignore"))
}

func xdgConfigGitCommitSummaryIgnore() string {
	config, err := xdg.SearchConfigFile("git-commit-summary/.gitcommitsummaryignore")
	if err != nil {
		return ""
	}
	return readFileString(config)
}

func validateIgnorePattern(pattern string) bool {
	if strings.ContainsRune(pattern, 0) {
		return false
	}
	if strings.HasPrefix(pattern, "!") {
		return false
	}
	if strings.Contains(pattern, "\\") {
		return false
	}

	_, err := path.Match(pattern, pattern)
	return err == nil
}

func readFileString(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}
