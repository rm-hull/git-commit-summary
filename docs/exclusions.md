# Excluding files from summary

The `git-commit-summary` tool supports excluding specific files or patterns from the generated commit summaries. This is useful for ignoring generated files, lockfiles, or sensitive data that shouldn't be part of the summary analysis.

Additionally, the tool supports automatic exclusion of files with excessive churn based on a configurable token limit, which is particularly useful for large diffs that might exceed LLM context windows.

## How it works

The tool searches for a file named `.gitcommitsummaryignore` in the following locations, merging the rules found:

1.  **Project Root:** Looks for `.gitcommitsummaryignore` starting from the current working directory and traversing up to the Git repository root.
2.  **User Home:** Looks for `~/.gitcommitsummaryignore`.
3.  **XDG Config:** Looks for `~/.config/git-commit-summary/.gitcommitsummaryignore` (respecting XDG Base Directory Specification).

## File Format

The `.gitcommitsummaryignore` file uses a simple line-based format:

*   Each line is a glob pattern to be ignored.
*   Lines starting with `#` or containing `//` are treated as comments and ignored.
*   Empty lines are ignored.
*   Only valid `path.Match` patterns are supported.
*   Patterns starting with `!` (negation) are currently not supported.
*   Backslashes `\` are not allowed.

## Example

```text
# Ignore dependency lock files
package-lock.json
yarn.lock
go.sum

# Ignore generated code
generated/*.go

# Ignore documentation images
docs/*.png
```

## Token Limit Filtering

To prevent large diffs from consuming excessive tokens or exceeding LLM context windows, `git-commit-summary` can automatically exclude files that have a high number of changed characters.

### Configuration

Set the `MAX_TOKENS` environment variable in your config file (`~/.config/git-commit-summary/config.env`) to specify the maximum number of tokens to spend on analyzing each individual file's diff. Files whose diff size exceeds this limit will be automatically excluded from the summary generation.

```env
# Exclude files with diffs larger than 2000 tokens (roughly 8000 characters)
MAX_TOKENS=2000
```

The token estimation uses a rough heuristic of approximately 4 characters per token. A value of `0` (the default) disables this filtering.

### How it works

1. The tool analyzes the staged diff and calculates the character count for added and removed lines for each file.
2. It estimates the token count using the formula: `tokens = ceil(characters / 4)`
3. Files with token counts exceeding the configured limit are automatically excluded from the diff sent to the LLM provider.

This is particularly useful for:

- **Large lockfile updates** (`package-lock.json`, `go.sum`, etc.)
- **Bulk code refactors** that touch many lines
- **Minified assets** or large generated files
