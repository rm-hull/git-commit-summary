# TODO: Go Code Improvements for git-commit-summary

This document outlines a plan to improve the `git-commit-summary` Go application.

## 1. Use `text/template` for the Prompt
The prompt is currently a simple string. Using the `text/template` package would make it more flexible and easier to maintain.

- **Create a new `prompt.tmpl` file:** This file will contain the prompt template.
- **Use the `text/template` package to parse and execute the template:** This will allow you to use variables and functions in the prompt.

## 2. Improve User Interaction
The `internal.TextArea` function is not very descriptive.

- **Rename `internal.TextArea` to `editCommitMessage`:** This will make the function's purpose more clear.
- **Improve the user interface for editing the commit message:** Consider if a more robust editor integration (like calling `vim` or `nano`) is necessary, or if the current TUI is sufficient.

## 3. Add a `Makefile`
A `Makefile` would automate common development tasks, such as building, testing, and running the application.

- **Create a `Makefile`:** This file will contain rules for building, testing, and running the application.
- **Add rules for common tasks:** The `Makefile` should include rules for `build`, `test`, `run`, and `clean`.

## 4. Error Handling
- **Provide more context for errors:** When an error occurs, log the error message and any relevant context, such as the file or line number where the error occurred.

## 5. Repository-Specific Configuration
Allow projects to define their own commit standards.

- **Support local configuration:** Check for `.git-commit-summary.yaml` in the current working directory.
- **Merge with global config:** Ensure local settings take precedence over global ones.

## 6. LLM Streaming Support
Improve perceived latency by streaming LLM output.

- **Integrate streaming:** Hook into `bubbletea` to render tokens as they arrive from the LLM provider.

## 7. Git Hook Integration
Automate the workflow for developers.

- **Create `install-hook` command:** Implement a flag or command that generates a `.git/hooks/prepare-commit-msg` script.
