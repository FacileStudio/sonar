# sonar

Multi-engine web search for the terminal. sonar runs one query across keyed
search APIs (Brave, Tavily, Exa, Firecrawl, SerpApi), your SearXNG instances,
and scrape fallbacks (Bing, DuckDuckGo), then merges the results into a single
ranked list. It is built to dodge IP-based blocking: engines are dispatched in
priority order with throttling and jitter, a per-engine circuit breaker parks
any backend that starts failing, and responses are cached so a repeated query
is not refetched.

## Engines

| Engine | Kind | Default |
| --- | --- | --- |
| brave | keyed API | off |
| tavily | keyed API | off |
| exa | keyed API | off |
| firecrawl | keyed API | off |
| serpapi | keyed API | off |
| searxng | SearXNG instances (from `searxng` config) | off (none configured) |
| bing | scrape | off |
| ddg | scrape | off |

No engine is enabled out of the box; turn on the keyed APIs you have keys for
and point `searxng` at your own instances.

## Usage

```sh
sonar search "circuit breaker pattern"
sonar search -n 5 "go 1.26 modules"
sonar search --json "lipgloss colors"
sonar open 3 "go 1.26 modules"   # open result 3 in the browser
sonar config                     # effective config: engines, keys, knobs
sonar cache                      # cached result set count and location
sonar cache clear                # empty the cache
```

`-n, --count N` caps the number of results (default: the config `count`, 10).
`--json` prints one JSON document on stdout and nothing else.

`sonar open [n] [query...]` runs a search and hands result `n` (default 1) to
the platform browser (`xdg-open` / `open` / `cmd start`). When no browser is
available it prints the title and URL instead.

Help pages, usage, errors, `--version` and shell completions are styled by
[fang](https://github.com/charmbracelet/fang), Charm's cobra starter kit.
`sonar man` emits a real roff man page and `sonar completion <shell>` prints a
completion script, both provided by fang.

Output colors are adaptive: result styling picks colors that read on a light or
dark terminal, and color (and ANSI) is stripped entirely when stdout is not a
terminal, so piping to a file or `jq` never leaks escape codes.

## MCP server

`sonar mcp` serves the search tool to an agent over MCP's stdio transport. An
agent launches the binary as a subprocess and speaks JSON-RPC over stdin and
stdout; nothing else is printed to the terminal.

For nacelle, add a `sonar` entry beside the others under `mcp:` in
`~/.nacelle.yml`:

```yaml
mcp:
  sonar:
    command: sonar
    args: [mcp]
```

The tool is `search(query, count)`: it runs the same dispatch, dedupe and
ranking as `sonar search`, so both surfaces return identical results for the
same query. Results are cached and throttled exactly as on the CLI. API keys
are read from the environment the same way the CLI reads them, which an agent
launched from a normal shell inherits — no extra setup.

## Configuration

`~/.sonar.yml`, or the path in `SONAR_CONFIG`. A missing file is not an error —
the defaults stand. Top-level keys: `count`, `timeout`, `minInterval`,
`jitter`, `breakerCooldown`, `cacheDir`, `cacheTTL`, `searxng`,
`searxngPriority`, and `engines.*` with `enabled`, `apiKey`, `baseURL`,
`priority`, `count` per engine.

Point sonar at the SearXNG instances you run with the `searxng` list — one URL
per instance. Each becomes its own engine, dispatched at `searxngPriority`
(default 9, so after the keyed APIs and before the scrapers). Leave it empty to
disable SearXNG entirely; none is bundled with sonar.

```yaml
searxng:
  - https://search.example.com
  - https://mirror.example.net
searxngPriority: 9
```

API keys belong in the file or in the environment: `SONAR_<ENGINE>_KEY`
(`SONAR_BRAVE_KEY`, `SONAR_TAVILY_KEY`, ...). The common provider names are
read as environment aliases only: `EXA_KEY`, `FIRECRAWL_KEY`, `TAVILY_KEY`,
`SERPAPI_API_KEY`. sonar also reads keys and the Bright Data zone from
[tiroir](https://github.com/FacileStudio/tiroir-cli) under the same
`SONAR_<ENGINE>_KEY` / `SONAR_BRIGHTDATA_ZONE` names, which is what makes the
MCP server work from a bare spawn — it inherits no shell exports, so without
the tiroir fallback it would silently skip every keyed engine. Precedence:
config file, then environment, then tiroir.

## Exit codes

`0` success. `1` error — a failing search, or a missing query.