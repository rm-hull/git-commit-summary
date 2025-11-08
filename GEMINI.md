# Project Overview

This project is a Go application that automatically generates commit summaries using a Large Language Model (LLM). It analyzes the staged changes in a Git repository and creates a concise and informative commit message. It currently supports both the Google Gemini and OpenAI APIs.

## Main Technologies

-   **Go:** The application is written in the Go programming language.
-   **Gemini API:** It uses the Google Gemini API to generate the commit summaries.
-   **OpenAI API:** It uses the OpenAI API to generate the commit summaries.
-   **Cobra:** For creating a powerful and modern CLI application.
-   **Godotenv:** For managing environment variables.
-   **adrg/xdg:** For XDG Base Directory Specification compliance.
-   **VersionInfo:** To provide version information.

# Building and Running

## Prerequisites

-   Go 1.25 or higher
-   A valid API key for your chosen LLM provider (Gemini or OpenAI).

## Building

To build the application, run the following command:

```bash
go build
```

## Running

`git-commit-summary` is XDG compliant, meaning it looks for its configuration file in a standard location. Create a `config.env` file in your XDG config home directory (e.g., `~/.config/git-commit-summary/config.env`) and add your configuration.

### Provider Configuration

You can select your provider by setting the `LLM_PROVIDER` environment variable. Supported values are `google` (default) and `openai`.

#### Google Gemini

Set the following in your `config.env` file:

```
LLM_PROVIDER=google
GEMINI_API_KEY=<your_api_key>
```

Get a Gemini API key from: https://aistudio.google.com/api-keys

You can also optionally set the `GEMINI_MODEL` environment variable to specify which model to use. The default is `gemini-2.5-flash-preview-09-2025`. **Note:** This default should be maintained to ensure consistency.

#### OpenAI

Set the following in your `config.env` file:

```
LLM_PROVIDER=openai
OPENAI_API_KEY=<your_api_key>
```

You can also optionally set `OPENAI_MODEL` (default: `gpt-4o`) and `OPENAI_BASE_URL`.

### Local Overrides

For local development or repository-specific overrides, you can still create a `.env` file in the project root. The application loads the local `.env` file *after* the global XDG configuration, so any variables in your local `.env` file will correctly override the global settings.

For more information on the XDG Base Directory Specification, see: [https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html)

### Run the application

```bash
./git-commit-summary
```

### Flags

| Flag             | Shorthand | Description                                                  |
| ---------------- | --------- | ------------------------------------------------------------ |
| `--version`      | `-v`      | Display version information                                  |
| `--message`      | `-m`      | Append a message to the commit summary                       |
| `--llm-provider` |           | Use specific LLM provider, overrides `LLM_PROVIDER` environment variable |

# Development Conventions

-   **Code Style:** The project follows the standard Go formatting guidelines. Use `gofmt` to format your code. It is also recommended to run `golangci-lint` to ensure code quality and consistency. Pay special attention to checking errors on `defer` statements (e.g., `defer reader.Close()`), as the linter will flag unchecked errors.
-   **Dependencies:** Dependencies are managed using Go modules. Use `go get` to add new dependencies and `go mod tidy` to clean up unused ones.
-   **Commits:** Commit messages should be concise and descriptive.
-   **Comments:** Comments should be used sparingly. Focus on *why* something is done, not *what* is done. Avoid adding comments that explain obvious code.
