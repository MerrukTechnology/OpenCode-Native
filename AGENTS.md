# OpenCode Development Guide

This document defines how agentic coding agents should operate in this repository, including build, lint, test commands, and code style guidelines. It supersedes older versions.

## Build/Lint/Test

- Build: `./scripts/snapshot` (uses goreleaser)
- Test: `go test ./...` (all packages) or `go test ./internal/llm/agent` (single package)

- Running a single test:
  - In a specific package: `go test -run '^TestMyFunction$' ./path/to/package -v`
  - Cross-package pattern: `go test -run '^(TestFoo|TestBar)$' ./... -v`

- Final checks: `make test` (runs all tests and formatters)
- Generate schema: `go run cmd/schema/main.go > opencode-schema.json`
- Generate mocks: `go generate ./...`
- DB migrations and SQL generation: see `internal/db/sql/` and run `sqlc generate`
- Security check: `./scripts/check_hidden_chars.sh`

## Linting and Style

- Formatting: `go fmt ./...` and `gofmt -w .` (or `goimports -w .` if available)
- Vet: `go vet ./...`
- Lint: `golangci-lint run` (preferred); fallback: `staticcheck ./...`
- Dependency tidy: `go mod tidy`

## Code Style Guidelines

### Imports
- Three groups: standard library, external, internal
- Separate groups with blank lines
- Sort each group alphabetically
- Internal imports use the repo module path, e.g. `github.com/MerrukTechnology/OpenCode-Native/internal/...`

### Naming
- Variables: camelCase (e.g., `filePath`, `contextWindow`)
- Functions: exported names in PascalCase; unexported in camelCase
- Types/Interfaces: PascalCase; interfaces often end with "Service"
- Packages: lowercase, single word (e.g., `agent`, `config`)

### Error Handling
- Return errors early; avoid deep nesting
- Wrap with context using `%w`: `fmt.Errorf("context: %w", err)`
- Prefer sentinel errors only for well-defined, reusable conditions

### Testing
- Table-driven tests with anonymous structs
- Subtests via `t.Run(...)`
- Test function names: `Test<Something>` and align with behavior
- Generate mocks with `mockgen` and place under `<pkg>/mocks/`

### Formatting and Tools
- Run `go fmt ./...` as part of CI or pre-commit
- Use `goimports` or configure your editor to fix imports automatically
- Enable and configure `golangci-lint` for a single, unified lint pass

### Documentation
- Exported API must have doc comments
- Write concise inline comments only for non-obvious logic

### Performance
- Minimize allocations and avoid allocations in hot paths
- Prefer `sync.Pool` or pooling when appropriate
- Measure before optimizing; ensure changes are beneficial

## Cursor Rules

- Cursor rules: none detected in this repository

## Copilot Instructions

- Copilot: follow project guidelines; no special restrictions beyond that

End of AGENTS.md
