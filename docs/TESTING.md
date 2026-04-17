# Testing Strategy

CopyTool uses layered testing focused on regression-prone behavior.

## Test Scope

1. Unit tests for core selection/filter/export logic (`internal/core`)
2. Path resolution and output target behavior
3. Gitignore and directory filtering behavior in tree scanning

## Required Commands

Run before opening a pull request:

```bash
go test ./...
go test -race ./...
go vet ./...
```

Generate local coverage report:

```bash
go test -covermode=atomic -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

## Determinism Rules

Tests should:

- use `t.TempDir()` for filesystem interactions
- avoid depending on local machine state
- avoid network calls
- avoid global mutable state unless reset in cleanup

## CI Behavior

CI executes tests across Linux, macOS, and Windows to catch platform-specific path/IO regressions.

The CI coverage job enforces a minimum total coverage threshold of `10.0%`.
