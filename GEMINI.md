# Project Overview

This project is a Go application that automatically generates commit summaries using the Gemini API. It analyzes the staged changes in a Git repository and creates a concise and informative commit message.

## Main Technologies

-   **Go:** The application is written in the Go programming language.
-   **Gemini API:** It uses the Google Gemini API to generate the commit summaries.
-   **Cobra:** For creating a powerful and modern CLI application.
-   **Gookit/Color:** For providing colorful output in the terminal.
-   **Spinner:** To display a loading spinner while generating the summary.
-   **Godotenv:** For managing environment variables.
-   **adrg/xdg:** For XDG Base Directory Specification compliance.
-   **VersionInfo:** To provide version information.

# Building and Running

## Prerequisites

-   Go 1.25 or higher
-   A valid API key for the Gemini API

## Building

To build the application, run the following command:

```bash
go build
```

## Running

`git-commit-summary` is XDG compliant, meaning it looks for its configuration file in a standard location. Create a `config.env` file in your XDG config home directory (e.g., `~/.config/git-commit-summary/config.env`) and add your Gemini API key:

```
GEMINI_API_KEY=<your_api_key>
```

Get an API key from: https://aistudio.google.com/api-keys

You can also optionally set the `GEMINI_MODEL` environment variable to specify which model to use. The default is `gemini-2.5-flash-preview-09-2025`.

For local development or repository-specific overrides, you can still create a `.env` file in the project root.

For more information on the XDG Base Directory Specification, see: [https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html](https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html)

2. Run the application:

    ```bash
    ./git-commit-summary
    ```

### Flags

| Flag        | Shorthand | Description                            |
| ----------- | --------- | -------------------------------------- |
| `--version` | `-v`      | Display version information            |
| `--message` | `-m`      | Append a message to the commit summary |

# Development Conventions

-   **Code Style:** The project follows the standard Go formatting guidelines. Use `gofmt` to format your code.
-   **Dependencies:** Dependencies are managed using Go modules. Use `go get` to add new dependencies and `go mod tidy` to clean up unused ones.
-   **Commits:** Commit messages should be concise and descriptive.
