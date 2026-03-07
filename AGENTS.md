# OpenCode Development Guide

This document defines how agentic coding agents should operate in this repository, including build, lint, test commands, and code style guidelines. It supersedes older versions.

## Quick Reference Commands

### Build & Development
- **Build**: `./scripts/snapshot` (uses goreleaser)
- **Test**: `go test ./...` (all packages) or `go test ./internal/llm/agent` (single package)
- **Final checks**: `make test` (runs all tests and formatters)
- **Generate schema**: `go run cmd/schema/main.go > opencode-schema.json`
- **Generate mocks**: `go generate ./...`
- **DB migrations**: See `internal/db/sql/` and run `sqlc generate`
- **Security check**: `./scripts/check_hidden_chars.sh`

### Code Quality
- **Formatting**: `go fmt ./...` and `gofmt -w .` (or `goimports -w .` if available)
- **Vet**: `go vet ./...`
- **Lint**: `golangci-lint run` (preferred); fallback: `staticcheck ./...`
- **Dependency tidy**: `go mod tidy`

## Testing Guide

### Running Tests
- **Single test**: `go test -run '^TestMyFunction$' ./path/to/package -v`
- **Cross-package pattern**: `go test -run '^(TestFoo|TestBar)$' ./... -v`

### Test Structure
- Use table-driven tests with anonymous structs
- Use subtests via `t.Run(...)`
- Test function names: `Test<Something>` and align with behavior
- Generate mocks with `mockgen` and place under `<pkg>/mocks/`

## Code Style Guidelines

### Imports
- Three groups: standard library, external, internal
- Separate groups with blank lines
- Sort each group alphabetically
- Internal imports use the repo module path, e.g. `github.com/MerrukTechnology/OpenCode-Native/internal/...`

### Naming Conventions
- **Variables**: camelCase (e.g., `filePath`, `contextWindow`)
- **Functions**: exported names in PascalCase; unexported in camelCase
- **Types/Interfaces**: PascalCase; interfaces often end with "Service"
- **Packages**: lowercase, single word (e.g., `agent`, `config`)

### Error Handling
- Return errors early; avoid deep nesting
- Wrap with context using `%w`: `fmt.Errorf("context: %w", err)`
- Prefer sentinel errors only for well-defined, reusable conditions

### Documentation
- Exported API must have doc comments
- Write concise inline comments only for non-obvious logic

### Performance
- Minimize allocations and avoid allocations in hot paths
- Prefer `sync.Pool` or pooling when appropriate
- Measure before optimizing; ensure changes are beneficial

## Agent Configuration

### Agent Fields Reference

| Field | Type | Description |
|-------|------|-------------|
| `model` | string | Model ID to use for this agent |
| `maxTokens` | number | Maximum response tokens |
| `reasoningEffort` | string | For models that support it (`low`/`medium`/`high`) |
| `mode` | string | `agent` (primary, switchable via tab) or `subagent` (invoked via task tool) |
| `name` | string | Display name for the agent |
| `description` | string | Short description of agent's purpose |
| `prompt` | string | Custom system prompt (overrides builtin prompt) |
| `color` | string | Badge color for subagent indication in TUI (e.g., `primary`, `secondary`, `warning`, `error`, `info`, `success`) |
| `hidden` | boolean | If true, agent is not shown in TUI switcher or subagent lists |
| `native` | boolean | Whether this is a built-in agent (set automatically) |
| `permission` | object | Agent-specific permission overrides |
| `tools` | object | Enable/disable specific tools (e.g., `{"skill": false, "bash": false}`) |

### Built-in Agents

| Agent | Purpose | Tools |
|-------|---------|-------|
| `coder` | Main coding agent, can spawn subagents | All tools |
| `hivemind` | Supervisory agent, coordinates subagents | Read-only tools |
| `explorer` | Codebase exploration subagent | Read-only tools |
| `workhorse` | Autonomous coding subagent | All tools |
| `summarizer` | Summarization subagent | No tools |
| `descriptor` | Short description generation subagent | No tools |

### Custom Agent Configuration

Agents can be defined in `.opencode.json` or as markdown files with YAML frontmatter.

**JSON Configuration:**
```json
{
  "agents": {
    "coder": {
      "model": "vertexai.claude-sonnet-4-5-m",
      "maxTokens": 64000,
      "reasoningEffort": "medium",
      "permission": {
        "skill": {
          "internal-*": "allow",
          "experimental-*": "deny"
        }
      },
      "tools": {
        "skill": true
      }
    }
  }
}
```

**Markdown Configuration:**
```markdown
---
name: Code Reviewer
description: Reviews code for quality, security, and best practices
mode: subagent
color: info
permission:
  bash:
    "*": deny
  edit:
    "*": deny
tools:
  bash: false
  write: false
---

You are a code review specialist...
```

### Agent Discovery Order

Agents are discovered from these locations in priority order (lowest to highest):
1. `~/.config/opencode/agents/*.md` (global)
2. `~/.agents/types/*.md` (global)
3. `.opencode/agents/*.md` (project)
4. `.agents/types/*.md` (project)
5. `.opencode.json` `agents` config (project - highest priority)

## Skills System

Skills are reusable instruction sets that agents can load on-demand.

**Key Concepts:**
- Skills are markdown files with YAML frontmatter
- Discovered from `.opencode/skills/`, `.agents/skills/`, `~/.config/opencode/skills/`, `~/.agents/skills/`, and custom paths
- Permissions control which skills agents can access
- Agent-specific permissions override global permissions

**Permission Patterns:**
- Exact match: `git-release: allow`
- Wildcards: `internal-*: deny`, `*-test: ask`
- Global: `*: ask`

## Permission System

Permissions use pattern matching with priority:

1. **Agent tool disable**: `agents.coder.tools.bash = false` → deny
2. **Agent-specific**: `agents.coder.permission.bash.{"git *": "allow"}`
3. **Global**: `permission.rules.bash = "ask"` or `permission.skill.internal-* = deny`
4. **Default**: ask

**Actions:**
- `allow`: Execute immediately
- `deny`: Block access
- `ask`: Prompt user (default)

**Granular Permissions:**
```json
{
  "permission": {
    "skill": { "*": "ask", "internal-*": "allow" },
    "rules": {
      "bash": { "*": "ask", "git *": "allow", "rm -rf *": "deny" },
      "edit": { "*": "allow", "*.env": "deny" },
      "read": { "*": "allow" },
      "task": { "*": "allow", "explorer": "allow" }
    }
  }
}
```

**Supported Permission Keys:**

| Key | Granular Pattern | Example |
|-----|-----------------|---------|
| `skill` | Skill name glob | `{"internal-*": "allow", "*": "ask"}` |
| `bash` | Command glob | `{"*": "ask", "git *": "allow"}` |
| `edit` | File path glob | `{"*": "deny", "src/**/*.go": "allow"}` |
| `read` | File path glob | `{"*": "allow", "*.env": "deny"}` |
| `task` | Subagent name glob | `{"*": "allow", "explorer": "allow"}` |

## TUI Usage

### Agent Switching
- Press `tab` to cycle through primary agents (mode=`agent`, hidden=false)
- The active agent is shown in the status bar
- Agent switching applies to the next new session

## Security & Best Practices

### Security Checks
- Always run `./scripts/check_hidden_chars.sh` before commits
- Never commit secrets or credentials
- Use the security check as part of your pre-commit workflow

### Development Workflow
1. Make changes to code
2. Run `go fmt ./...` and `golangci-lint run`
3. Run tests: `go test ./...`
4. Run security check: `./scripts/check_hidden_chars.sh`
5. Run final checks: `make test`

## Troubleshooting

### Common Issues
- **Build failures**: Check `go.mod` with `go mod tidy`
- **Test failures**: Use `go test -v` for verbose output
- **Lint errors**: Run `golangci-lint run --fix` if available
- **Permission denied**: Check file permissions and agent configuration

### Getting Help
- Check the project README for additional guidelines
- Look at existing code patterns for style consistency
- Use the explorer agent for codebase navigation

## Technical Debt / Follow-up Issues

### HIGH Priority
- [ ] Fix unchecked error returns (~140 errcheck issues)
- [ ] Add context.Context to exec.Command calls (~18 locations)

### MEDIUM Priority
- [ ] Refactor strings.Index → strings.Cut (7 files)
- [ ] Add package doc comments (6 files)

### Completed
- [x] Context propagation for critical DB/goroutine operations (commit 8616e87)