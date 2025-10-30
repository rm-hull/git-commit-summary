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

## Installation

### Set up your API key

`git-commit-summary` is XDG compliant, meaning it looks for its configuration file in a standard location. Create a `config.env` file in your XDG config home directory (e.g., `~/.config/git-commit-summary/config.env` on Linux, `~/Library/Application Support/git-config-summary/config.env` on macOS, or `%USERPROFILE%\.config\git-commit-summary\config.env` on Windows).

You can configure the LLM provider by setting the `LLM_PROVIDER` environment variable. The supported providers are `google` (default) and `openai`.

For local development or repository-specific overrides, you can still create a `.env` file in your git repository root.

For more information on the XDG Base Directory Specification, see: [https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html)

#### Google

Add your Gemini API key to the `config.env` file:

```
GEMINI_API_KEY=<your_api_key>
```

Get an API key from: https://aistudio.google.com/api-keys

You can also optionally set the `GEMINI_MODEL` environment variable to specify which model to use. The default is `gemini-2.5-flash-preview-09-2025`.

#### OpenAI

Add your OpenAI API key to the `config.env` file:

```
OPENAI_API_KEY=<your_api_key>
```

You can also optionally set the `OPENAI_MODEL` environment variable to specify which model to se. Note, if no model is selected, `gpt-4o` is used as the default.

Additionally, you can also set `OPENAI_BASE_URL` - this is especially useful if you want to point to a local [llama.cpp](https://github.com/ggml-org/llama.cpp) service. For example, you may use the following config:

```
LLM_PROVIDER="openai"
OPENAI_BASE_URL="http://localhost:8080/v1"
OPENAI_MODEL="Meta-Llama-3.1-8B-Instruct-Q4_K_M"
```

## Usage

Once installed, check that the executable is on the $PATH, with `git-commit-summary --version`. Then, as part of your development workflow

1.  **Stage your changes:**

    ```bash
    git add <files>
    ```

2.  **Run the tool:**

    ```bash
    git commit-summary
    ```

3.  **Confirm the commit:**

    The tool will display the generated commit summary and ask for your confirmation. Type `y` to accept and commit, or `n` to abort.

## Flags

| Flag             | Shorthand | Description                                                                                                                                      |
| ---------------- | --------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| `--version`      | `-v`      | Display version information                                                                                                                      |
| `--message`      | `-m`      | Append a message to the commit summary                                                                                                           |
| `--llm-provider` | _n/a_     | Use the specific LLM provider: supported values are currently only **google** & **openai**. Overrides the `LLM_PROVIDER` environmental variable. |

## Aliases

If you want to use a shorter command, you can add an alias to your `~/.gitconfig` file. For example, to use `git cs` as a shorthand for `git commit-summary`, you can run the following command:

```bash
git config --global alias.cs commit-summary
```

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.
