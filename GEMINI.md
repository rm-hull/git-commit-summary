# Project Overview

This project is a Go application that automatically generates commit summaries using the Gemini API. It analyzes the staged changes in a Git repository and creates a concise and informative commit message.

## Main Technologies

*   **Go:** The application is written in the Go programming language.
*   **Gemini API:** It uses the Google Gemini API to generate the commit summaries.
*   **Gookit/Color:** For providing colorful output in the terminal.
*   **Spinner:** To display a loading spinner while generating the summary.
*   **Godotenv:** For managing environment variables.

# Building and Running

## Prerequisites

*   Go 1.25 or higher
*   A valid API key for the Gemini API

## Building

To build the application, run the following command:

```bash
go build
```

## Running

1.  Create a `.env` file in the project root and add your Gemini API key:

    ```
    GEMINI_API_KEY=<your_api_key>
    ```

2.  Run the application:

    ```bash
    ./git-commit-summary
    ```

# Development Conventions

*   **Code Style:** The project follows the standard Go formatting guidelines. Use `gofmt` to format your code.
*   **Dependencies:** Dependencies are managed using Go modules. Use `go get` to add new dependencies and `go mod tidy` to clean up unused ones.
*   **Commits:** Commit messages should be concise and descriptive.
