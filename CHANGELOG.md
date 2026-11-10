# Changelog

All notable changes are documented here. Format derives from Keep a Changelog.

## [Unreleased]

### Added

- `POST /mcp`, the remote MCP server (spec 2026-07-28, stateless, tronc/mcp), behind Registre tokens bound to `https://sonar.facile.studio/mcp` with the one scope `sonar.read`. Two read-only, open-world tools: `sonar_search` (web, images or news; `engines` only on the user's own keys; no `!bang`) and `sonar_fetch` (public-only guard, markdown in parts with `offset` and `next_offset`). Each call spends one search of the daily quota; an exhausted quota is an error naming the upgrade page. On when `MCP_RESOURCE` is set (compose default).

### Changed

- tronc v0.20.0, porte v0.14.0 with porte/mcpauth v0.1.0, go-sdk v1.8.0 (the stdio `sonar mcp` included).

## [v0.15.6] - 2026-11-10

### Changed

- Hosted web searches call Brave, then Tavily only when Brave returns fewer than 5 results, errors or times out: 1 or 2 engines per search instead of every keyed engine. `SONAR_WEB_ENGINES`, `SONAR_WEB_ENOUGH` set the policy. CLI and library defaults are unchanged.
- The AI summary takes only the query and finds its own sources on Tavily, Exa or Parallel (`SONAR_SUMMARY_ENGINES`), never Brave, and returns the sources it cites. It stays off until Bulle serves `POST /api/ai/summarize`.
- The "Sonar never stores what you search" note now sits in Settings, Preferences.

### Fixed

- `TestGuardedBrowserRefusesInternalNavigation` holds the websocket page's load until the refusal happened, so it no longer races on slow runners.

## [v0.15.5] - 2026-11-08

### Fixed

- Optically aligned the search icon with placeholder text and relaxed icon-to-text spacing inside `SearchBox`.

## [v0.15.4] - 2026-11-08

### Fixed

- Restored full width on the home search bar by explicitly setting `w-full` on `SearchBox` and its container.

### Removed

- Removed the privacy note below the search input on the home page.

## [v0.15.3] - 2026-11-08

### Changed

- Moved the tagline ("Private web search. No history, no tracking.") directly under the search input bar on the home page.

## [v0.15.2] - 2026-11-08

### Added

- Settings sections use Solar icons from `@facile/muse`.
- Monochrome provider logos in the engine keys settings section.

### Changed

- Removed duplicate center logo and heading from the home search page.

### Fixed

- Authenticated users now display their Registre profile picture: OIDC picture claims are stored at sign-in, empty rows are backfilled from identities at boot, and sessions resolve identity avatar URLs.

## [v0.15.1] - 2026-11-08

### Fixed

- A search whose caller went away (a closed tab, an aborted request) is refunded: the refund no longer runs on the cancelled request context, and a failed refund logs its cause.

## [v0.15.0] - 2026-11-08

### Added

- `sonar serve` runs the hosted sonar at sonar.facile.studio: sign-in with a Facile account through Registre, the JSON API (`/api/search` for web, images and news, `/api/fetch`, `/api/suggest`, `/api/usage`, `/api/bangs`), a signed and guarded image proxy at `/img`, an OpenSearch description at `/opensearch.xml` and the API reference at `/docs`.
- Daily search quotas per person from the `sonar.searches_daily` plan entitlement (free plan 20 until Ardoise publishes values). An exhausted quota answers 402 with an upgrade link; a search no engine answers is refunded.
- People can bring their own engine API keys, sealed at rest with porte's sealer; searches on them skip the quota.
- `!bang` shortcuts: internal ones pick a tab or an engine, external ones (`!g`, `!ddg`, `!w`, ...) redirect, with a privacy note.
- A SvelteKit web client on muse: web, images and news tabs, keyboard navigation (`/`, arrows, `j`/`k`), typing suggestions that can be turned off, settings for theme, language and own keys, and a draft of the terms of use.
- The paid AI summary through Bulle, gated on `ai.access`, ships turned off (`SONAR_AI_SUMMARY`) until Bulle serves the route described in docs/ai-summary.md.
- `websearch.Client.Images` and `websearch.Client.News` (Brave).

### Changed

- Brave results are never cached, in the CLI, the library or the server, as Brave's API terms require.
- Brave titles and snippets come back as plain text, without the `<strong>` highlights and HTML entities.

### Privacy

- No query reaches a log: the request logger redacts `q`, `u`, `s` and `url`, and the compose file turns Traefik access logs off for sonar's routes.

## [v0.14.1] - 2026-11-08

### Fixed

- `sonar fetch` (and `websearch.Fetch` with `Browser`) now uses an installed Chromium found on PATH or in the usual locations. Before, rod always downloaded its own glibc build, so on Alpine images with `apk add chromium` the browser path failed with `fork/exec ... chrome: no such file or directory`.

## [v0.14.0] - 2026-11-08

### Added

- `websearch.Client.Fetch(ctx, url, FetchOptions)` fetches one page as markdown from Go and returns a `*Page` with `FinalURL`, `Status`, `Title`, `Markdown`, `Engine` and `Truncated`. The `--public-only` guard is on by default (every redirect hop, every browser subrequest, dialer-level address check); `FetchOptions.AllowInternal` opts out for trusted callers. A refusal is a `*websearch.BlockedError` with a code. The plain HTTP path needs no Chromium; `Browser` opts into rendering and `Selector` needs it.

### Fixed

- A guarded browser fetch re-checks the guard after extraction, so a late redirect (such as a meta refresh that fires after load) that was refused fails the URL instead of returning the refusal page.

## [v0.13.1] - 2026-11-08

### Security

- `sonar fetch --public-only` (also `fetch.publicOnly: true` in the config or `SONAR_FETCH_PUBLIC_ONLY=1`) refuses any request that would reach an internal address. Before, a public URL that redirected (3xx, JavaScript `location`, meta refresh) or embedded iframes and subresources pointing at 127.0.0.1, 169.254.169.254, a Docker gateway or a container name was fetched and its content returned. Chromium now runs behind a loopback proxy in sonar that checks every request, redirect hop and websocket and connects only to the address it checked; the plain HTTP path dials the checked address, re-checks it in the dialer and re-checks every redirect. The MCP `fetch` tool honors the config and environment setting.

### Added

- `sonar fetch --json`: one object per URL with `url`, `title`, `markdown`, `engine`, or an `error` with a `code` (`blocked_url`, `blocked_internal`, `blocked_unresolved`, `fetch_failed`).
- `sonar fetch` prints each failed URL on stderr instead of dropping it silently.

## [v0.13.0] - 2026-11-03

### Added

- Public `websearch` Go package: `New(Config)` and `Search` run the multi-engine dispatcher as a library. Keys come only from `Config` (no config file, env or tiroir reads) and the cache can live in memory.

## [v0.12.1] - 2026-09-25

### Fixed

- Truncate search snippets to single-line summaries (up to 280 characters with ellipsis) and collapse whitespace to prevent multi-paragraph excerpt flooding in terminal results.
- Search fanout adds early grace window cutoff (1.2s) once enough results have arrived, terminating slow or hanging engine calls like unconfigured scrapers.
- Avoid terminal OSC background color query hang in CLI output by using `COLORFGBG` detection with dark-mode fallback.

## [v0.12.0] - 2026-09-25

### Added

- Parallel.ai engine (`parallel`): fast LLM-optimized search and extraction, reads `SONAR_PARALLEL_KEY`, `PARALLEL_API_KEY`, or tiroir `PARALLEL_API_KEY`.
- Multi-URL parallel `sonar fetch <URL>...`: fetches multiple web pages concurrently.
- Fast extraction pipeline: uses Parallel extract and fast HTTP fetch before falling back to headless browser.
- `--browser` (`-b`) flag on `sonar fetch` to force headless browser rendering when needed.
- `--timeout` (`-t`) flag on `sonar search` to override default engine search timeout.

### Changed

- Multi-engine dispatch acquires concurrency slots inside worker goroutines, preventing head-of-line blocking across keyed engines.

## [v0.11.1] - 2026-09-22

### Changed

- `sonar search` no longer animates a spinner on stderr. The braille
  frames duplicated lines on some terminals (a tty or line-ending mismatch),
  so the wait now runs silently.

## [v0.10.0]

### Added

- `sonar fetch <URL>` command: extracts web page content using a headless
  browser and prints the text to stdout. Supports `--wait-until` conditions
  (`networkidle`, `domcontentloaded`, `load`, `none`).
- Headless browser engine (`rod`) using `go-rod` with stealth mode and safe
  DOM element extraction.

### Fixed

- Multi-engine fan-out uses `sync.WaitGroup` so that transient errors or
  rate limits on one engine do not abort other concurrent engine queries.
- Clean error returns instead of panics during browser launches or element
  lookups.

## [v0.9.0]

### Added

- Per-call engine selection: `sonar search --engines tavily,exa` and the MCP
  tool's optional `engines` list run a named subset of the enabled engines,
  leaving the default fan-out to every enabled engine untouched. An unknown or
  disabled name fails with the list of usable engines.

### Changed

- The MCP server's duplicate search handlers share one path, so the
  dispatcher-backed and per-call fallbacks cannot drift.
## [v0.8.1]

### Fixed

- The singleflight key now includes the result count, so two concurrent
  queries with the same words but different `count` values each run their
  own engine call instead of sharing one.

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