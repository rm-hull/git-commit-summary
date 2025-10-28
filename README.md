# Git Commit Summary

Tired of writing git commit messages? This tool uses AI ✨ to automatically generate a concise commit summary for your staged changes.

## Features

-   **Automatic Commit Summaries:** Analyzes your staged changes and generates ~~AI-slop~~ high-quality commit messages.
-   **Interactive Confirmation:** Prompts you to confirm the commit message before committing.
-   **Colorful Output:** Provides a visually appealing and easy-to-read output in your terminal.

## Installation

```bash
go install github.com/rm-hull/git-commit-summary
```

## Usage

1.  **Set up your API key:**

    Create a `.env` file in your git repository and add your Gemini API key:

    ```
    GEMINI_API_KEY=<your_api_key>
    ```

    Get an API key from: https://aistudio.google.com/api-keys

    You can also optionally set the `GEMINI_MODEL` environment variable to specify which model to use. The default is `gemini-2.5-flash-preview-09-2025`.

2.  **Stage your changes:**

    ```bash
    git add <files>
    ```

3.  **Run the tool:**

    ```bash
    git commit-summary
    ```

4.  **Confirm the commit:**

    The tool will display the generated commit summary and ask for your confirmation. Type `y` to accept and commit, or `n` to abort.

## Flags

| Flag        | Shorthand | Description                            |
| ----------- | --------- | -------------------------------------- |
| `--version` | `-v`      | Display version information            |
| `--message` | `-m`      | Append a message to the commit summary |

## Aliases

If you want to use a shorter command, you can add an alias to your `~/.gitconfig` file. For example, to use `git cs` as a shorthand for `git commit-summary`, you can run the following command:

```bash
git config --global alias.cs commit-summary
```

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.
