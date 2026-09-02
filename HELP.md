# Git Commit Summary

Generate a concise, high-quality commit message from staged changes using an LLM. The tool inspects your staged diff, infers the intent of the work, and turns it into a clean commit summary you can accept, edit, regenerate, or reject.

It is designed to reduce commit-message churn while keeping control in your hands. You can provide extra guidance with `--hint`, review the diff in the built-in viewer, and commit with a single confirmation step when everything looks right.

## Features

- AI-assisted commit summaries primarily driven by your staged changes, while also incorporating the tone and style of your recent commit history to keep messages consistent.
- Optional `--hint` guidance so you can steer the model toward the right scope, tone, or context for the change.
- Built-in setup wizard and configuration support for choosing an LLM provider, model, and API key without manual editing.
- Bash and Zsh completion generation for faster shell usage and discoverability of supported flags.
- Git hook installation and removal so you can automate summary generation as part of the normal commit workflow.
- Interactive commit editor with approve, regenerate, abort, and preview actions to keep the final message under your control.
- Diff view for reviewing staged content directly in the terminal before committing.

## Usage

```bash
git-commit-summary [flags]
```

To make it easier to run in day-to-day work, you can set up a Git alias and invoke it as `git cs`:

```bash
git config --global alias.cs commit-summary
git cs
```

This alias calls the same CLI with the standard commit-summary workflow, so you can stage files and run `git cs` to generate and review the message in one step.

## Shortcuts in the commit editor

- `CTRL+A`: Accept and commit
- `CTRL+R`: Reprompt and regenerate the commit message
- `CTRL+K`: Clear the generated commit message
- `CTRL+X`: Cut the current line and copy it to the clipboard
- `CTRL+P`: Toggle preview mode
- `CTRL+D`: Open the raw colored diff; toggle between raw diff and compact summary while viewing the diff
- `ESC`: Return from the diff view, or abort the editor

## Project

Repository: https://github.com/rm-hull/git-commit-summary

This project is released under the MIT license, and contributions are welcome in the form of bug reports, feature ideas, and pull requests.

---