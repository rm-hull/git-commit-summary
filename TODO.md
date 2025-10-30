# TODO: Go Code Improvements for git-commit-summary

This document outlines a plan to refactor and improve the `git-commit-summary` Go application.

## 1. Refactor `main.go` and the `run` function

The `main.go` file currently contains a large `run` function that handles all the application logic. This makes it difficult to read, test, and maintain.

-   **Create a new `internal/app` package:** This package will encapsulate the core application logic.
-   **Move the `run` function's logic into a new `Run` function in `internal/app/app.go`:** This will separate the application logic from the command-line interface.
-   **Break down the `Run` function into smaller, more focused functions:** Each function should have a single responsibility, such as getting the git diff, generating the commit summary, or editing the commit message.

## 2. Improve Error Handling

Error handling is inconsistent. Some errors are logged and the application exits, while others are returned.

-   **Use a consistent error handling strategy:** Return errors from functions and handle them in the `main` function. This will make the code more robust and easier to debug.
-   **Provide more context for errors:** When an error occurs, log the error message and any relevant context, such as the file or line number where the error occurred.

## 3. Centralize Configuration

Configuration is currently scattered throughout the code.

-   **Create a new `internal/config` package:** This package will be responsible for loading and managing the application's configuration.
-   **Define a `Config` struct:** This struct will hold all the application's configuration settings.
-   **Load the configuration from a file or environment variables:** The `config` package should support loading the configuration from a variety of sources.

## 4. Use `text/template` for the Prompt

The prompt is currently a simple string. Using the `text/template` package would make it more flexible and easier to maintain.

-   **Create a new `prompt.tmpl` file:** This file will contain the prompt template.
-   **Use the `text/template` package to parse and execute the template:** This will allow you to use variables and functions in the prompt.

## 5. Improve User Interaction

The `internal.TextArea` function is not very descriptive.

-   **Rename `internal.TextArea` to `editCommitMessage`:** This will make the function's purpose more clear.
-   **Improve the user interface for editing the commit message:** Use a more user-friendly editor, such as `vim` or `nano`.

## 6. Add a `Makefile`

A `Makefile` would automate common development tasks, such as building, testing, and running the application.

-   **Create a `Makefile`:** This file will contain rules for building, testing, and running the application.
-   **Add rules for common tasks:** The `Makefile` should include rules for `build`, `test`, `run`, and `clean`.

## 7. Add Unit Tests

The application currently has no unit tests.

-   **Add unit tests for the core application logic:** This will help to ensure that the application is working correctly and prevent regressions.
-   **Use a testing framework, such as `testify`:** A testing framework will make it easier to write and run tests.
