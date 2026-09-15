# AGENTS.md

This is a Go CLI (`sonar`) — a multi-engine web search tool. It follows the Facile suite CLI conventions.

## Build & Test

```sh
go build ./...
go test ./...
filet check .
```

Go 1.26 is required (`go.mod`).

## Architecture

- `cmd/` — cobra commands (search, config, cache, mcp, open). Entry point in `main.go` calls `cmd.Execute()`.
- `internal/` — implementation, organized by concern (config, engines, search, merge, cache, httpc, politeness, mcpserver, tiroir).
- `internal/engines/` — one file per search engine (brave, tavily, exa, searxng, bing, ddg, etc.). Each engine implements a common result interface.

No external imports from `internal/` are allowed outside this module.

## Code Style & Standards

Enforced by `filet.yml` at the repo root. Key rules:
- Files <= 250 lines, functions <= 30 lines and <= 25 statements.
- Max 8 functions per file, 5 params, 3 returns, 4 nesting, complexity 10.
- No inline comments, no TODOs, no trailing whitespace.
- Doc comments required on public symbols.
- Run `filet check .` before committing; code that fails filet is invalid.

## Tooling

CLI framework: `charm.land/fang/v2` (cobra starter kit with enhanced output). Output styling uses `lipgloss/v2`.

MCP server: `internal/mcpserver/` serves search as an MCP stdio tool. The `sonar mcp` command is the entry point.

## Configuration

Config file: `~/.sonar.yml` (or `$SONAR_CONFIG`). Keys: `count`, `timeout`, `minInterval`, `jitter`, `breakerCooldown`, `cacheDir`, `cacheTTL`, `searxng`, `searxngPriority`, `engines.*` (enabled, apiKey, baseURL, priority, count per engine).

API keys: config file or `$SONAR_<ENGINE>_KEY` env vars. Provider aliases (`EXA_KEY`, `TAVILY_KEY`, etc.) and tiroir fallback are also read. Precedence: config file, environment, tiroir.

## Testing

- Live engine tests use `_live_test.go` suffix and require API keys; they do not run in CI without keys.
- `brightdata_test.go` covers the Bright Data scraper engine.
- Run `go test ./...` for all offline tests.
