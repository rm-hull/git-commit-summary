# AI Agent Handbook: `git-commit-summary`

This document defines the protocols, structure, and guidelines for an AI agent working on the `git-commit-summary` project.

## 1. Project Overview & Persona
- **Mission:** A Go CLI application that analyzes staged Git changes and generates commit summaries using LLMs (Gemini/OpenAI).
- **Architecture:** Cobra-based CLI, modularized Go (`/internal`), XDG-compliant configuration.
- **Agent Persona:** A precise, safety-conscious Go developer. Prioritize correctness, performance, and clean, idiomatic code over quick fixes.

## 2. Directory Structure
Agents should orient themselves using this map:
```text
.
├── main.go               # Entry point
├── internal/             # Core logic
│   ├── app/              # CLI command definitions (Cobra)
│   ├── config/           # XDG & env loading logic
│   ├── git/              # Git interaction wrappers
│   ├── llm_provider/     # LLM abstractions
│   ├── ui/               # Terminal output / formatting
│   └── setup/            # Initialization logic
├── docs/                 # Documentation
└── ...
```

## 3. Critical Coding Protocols
- **Error Handling:** 
    - **Must** check errors for all operations.
    - **Crucial:** Always check errors on `defer` statements (e.g., `defer f.Close()`). If the linter flags an unchecked error, it is a priority fix.
- **Formatting:** Always run `gofmt` after changes.
- **Dependencies:** Managed via Go modules (`go.mod`/`go.sum`). Use `go mod tidy` after modifying imports.
- **Comments:** Be sparse. Focus on *why* a complex logic branch exists, not *what* the code does.

## 4. Configuration & Environment
- **XDG Compliance:** Respect the XDG Base Directory Specification.
- **Configuration File:** Look in `~/.config/git-commit-summary/config.env` for production settings.
- **Local Overrides:** The application loads a local `.env` file *after* XDG. Agents should prioritize environment variables over hardcoded constants.
- **API Keys:** **NEVER** commit API keys or hardcode them in the codebase. Always reference them via environment variables (`GEMINI_API_KEY`, `OPENAI_API_KEY`).

## 5. Development & Testing Workflow
- **Building:** `go build`
- **Linting:** Run `golangci-lint` to ensure standards.
- **Running:** `./git-commit-summary [flags]`
- **Testing:** **Mandatory.** Every feature or fix must include unit tests in `internal/`. Run `go test ./...` before declaring a task complete. If tests are skipped, the task is not finished.

## 6. Known Flags (for integration/testing)
- `--version` / `-v`: Display version.
- `--message` / `-m`: Append extra text to the generated commit.
- `--llm-provider`: Override the `LLM_PROVIDER` environment variable.
