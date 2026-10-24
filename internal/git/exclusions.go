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

			excludes = append(excludes, ":(exclude)"+normalizeIgnorePattern(line))
		}
	}
	return dedupe(excludes)
}

// normalizeIgnorePattern converts .gitignore-style patterns to Git pathspec-compatible patterns.
// Git's :(exclude) pathspec does not support ** (recursive glob) or trailing-slash directory
// patterns the way .gitignore does. This function transforms:
//   - "build/" -> "**/build/**"
//   - "**/build/" -> "**/build/**"
//   - patterns ending with "/**" are left as-is
//   - file patterns without trailing "/" are left as-is
func normalizeIgnorePattern(pattern string) string {
	if !strings.HasSuffix(pattern, "/") {
		return pattern
	}

	// Remove trailing slash to get the directory path
	dirPath := strings.TrimSuffix(pattern, "/")

	// If it already ends with **, it's already a valid pathspec glob
	if strings.HasSuffix(dirPath, "**") {
		return pattern
	}

	// Strip leading **/ if present, then wrap with ** to match at any depth
	dirPath = strings.TrimPrefix(dirPath, "**/")
	return "**/" + dirPath + "/**"
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
