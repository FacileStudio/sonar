# Changelog

All notable changes are documented here. Format derives from Keep a Changelog.

## [Unreleased]

## [v0.8.0]

### Added

- `sonar search` shows a spinner on stderr while the engines run, so the
  wait is visible without polluting stdout or `--json` output.

## [v0.7.0]

### Added

- Parallel engine dispatch: keyed-API engines run concurrently, scrape
  engines (bing, ddg) stay sequential behind a capacity-1 semaphore so
  they never burst one IP.
- `maxConcurrency` (default 8) and `retryBackoff` (default 800ms) config
  knobs, shown in `sonar config`.
- Singleflight on the fetch path: concurrent identical queries share one
  engine call.

### Changed

- Each engine gets its own politeness throttle instead of one shared
  throttle serializing the whole walk; user-facing latency drops from the
  sum of engine times to the slowest engine.
- Search timeout budget is one engine timeout plus a 5s margin instead of
  twice the timeout.
- Ranking is deterministic across dispatch order: results carry their
  engine's priority and merge sorts on trust then priority.
- The MCP server builds its dispatcher once at startup, so per-engine
  circuit breakers persist across tool calls; a broken config falls back
  to the per-call path.

### Fixed

- Cache keys truncated long queries at 80 characters, so queries sharing
  a prefix served each other's cached results; keys are now a sha256 of
  the full query.
- The cache is mutex-guarded, making it safe under concurrent queries.

## [v0.6.0]

### Added

- `sonar install` registers the search MCP server in the agent harnesses on
  this machine: claude (~/.claude.json), gemini (~/.gemini/settings.json),
  opencode (~/.config/opencode/opencode.json) and nacelle-tui
  (~/.nacelle.yml sources.mcp). The merge is surgical for the YAML nacelle
  config so comments and other servers survive byte for byte, and refuses
  loudly (exit 1) on a config it cannot safely touch.