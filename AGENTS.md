# AGENTS.md

## Project Overview

`tainanfire` is a Telegram bot that monitors and broadcasts fire department incidents from Tainan, Kaohsiung, and Hsinchu. It scrapes data from official websites and sends updates to configured Telegram chats.

### Key Features
-   **Multi-Region Support**: Monitors Tainan, Kaohsiung, and Hsinchu fire departments.
-   **Real-time Updates**: Periodically fetches and filters incidents.
-   **Telegram Integration**: Broadcasts filtered events (e.g., fire incidents or major mobilizations) to specific chat groups.
-   **Dockerized**: Run easily with Docker Compose.

## Getting Started

### Prerequisites
-   **Go**: Version 1.25 or later.
-   **Docker**: Latest version with Compose V2 support.
-   **Environment Variables**:
    -   `API_KEY`: Telegram Bot API Key.

### Installation
1.  Clone the repository.
2.  Install dependencies:
    ```bash
    go mod download
    ```

### Running the Project
-   **Local Development**:
    ```bash
    export API_KEY="your_telegram_bot_token"
    go run .
    ```
-   **With Docker**:
    ```bash
    docker compose up --build
    ```

## Debugging

-   **Logs**: The application uses standard logging. Check stdout/stderr.
-   **Delve**:
    ```bash
    dlv debug .
    ```

## Build, Lint, and Test

### Build
```bash
go build -o fire .
```

### Lint
Use `golangci-lint` for linting.
```bash
golangci-lint run
```

### Test
Run all tests:
```bash
go test ./...
```
Run a specific test:
```bash
go test -v -run TestName ./path/to/package
```

## Code Style & Guidelines

### General Rules (GOD RULES)
1.  **Tone**: Say "meow" before answering or performing actions.
2.  **Git**:
    -   Never commit without clear instructions.
    -   Always sign commits (`-S`).
    -   Retry if signing fails.
3.  **Tools**:
    -   Use **Context7 MCP** for documentation/code generation/setup instructions.
    -   Use **pnpm** instead of npm (if Node.js is involved).

### Go Guidelines
1.  **Version**: Use Go 1.25.
2.  **Project Structure**:
    -   **Do not** use `pkg/`, `internal/`, or `cmd/`.
    -   Place main package files in the project root.
    -   Place sub-package directories in the project root (e.g., `bucket/`).
    -   Use monolithic architecture unless instructed otherwise.
3.  **Dependencies**:
    -   **Dependency Injection**: Use `github.com/samber/do/v2` if necessary.
    -   **Error Handling**: Use `github.com/samber/oops` for errors.
    -   **HTTP Framework**: Use `gin` (if implementing HTTP server).
    -   **Database**: Use `sqlc` + `sqlite` (if database is needed).
4.  **Syntax**:
    -   Use `for i := range n` (or `for range n`) instead of `for i := 0; i < n; i++`.

### Docker Guidelines
1.  **Compose File**:
    -   Name it `compose.yaml`. **Do not** use `docker-compose.yml` or `docker-compose.yaml`.
    -   Use `docker compose` command, not `docker-compose`.
    -   Do not include `version: '3.x'` in `compose.yaml`.

### Node.js Guidelines (If applicable)
1.  **Version**: Use Node.js v25.
2.  **Formatting**: Use Prettier with the specific configuration in `.prettierrc`.
3.  **Tailwind CSS**:
    -   Use v4.
    -   Import via `@import 'tailwindcss';` in CSS.
    -   Do not use `postcss` or `autoprefix`.
    -   Do not use `tailwind.config.js`.

## AI Agent Instructions
-   **Analysis**: Before making changes, analyze `go.mod`, `main.go`, and existing structures.
-   **Conventions**: Adhere to the guidelines above. If existing code violates them (e.g., `docker-compose.yml` exists instead of `compose.yaml`), prioritize migrating to the new standard unless explicitly told otherwise.
-   **Refactoring**: When touching legacy code, opportunistically refactor to match the "Go Guidelines" (e.g., switching `log` to `oops`).
