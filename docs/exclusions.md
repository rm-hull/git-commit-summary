# Excluding files from summary

The `git-commit-summary` tool supports excluding specific files or patterns from the generated commit summaries. This is useful for ignoring generated files, lockfiles, or sensitive data that shouldn't be part of the summary analysis.

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
