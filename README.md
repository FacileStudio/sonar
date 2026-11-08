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
| parallel | keyed API | off |
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
sonar search --engines tavily,searxng "go 1.26 modules"
sonar open 3 "go 1.26 modules"   # open result 3 in the browser
sonar config                     # effective config: engines, keys, knobs
sonar cache                      # cached result set count and location
sonar cache clear                # empty the cache
```

`-n, --count N` caps the number of results (default: the config `count`, 10).
`--engines a,b` runs only those engines, by config name — for example
`--engines tavily` for one keyed API or `--engines exa,searxng` for a subset.
A name that is not enabled fails with the list of usable engines. Omitted, the
search runs every enabled engine as usual.
`--json` prints one JSON document on stdout and nothing else.

`sonar open [n] [query...]` runs a search and hands result `n` (default 1) to
the platform browser (`xdg-open` / `open` / `cmd start`). When no browser is
available it prints the title and URL instead.

## Fetching pages

```sh
sonar fetch https://example.com
sonar fetch --json https://a.example https://b.example
sonar fetch --public-only --json "$untrusted_url"
```

`sonar fetch` prints each page as markdown. It tries a plain HTTP fetch first
and falls back to headless Chromium (`--browser` forces it). `--json` prints
one object per URL: `url`, `title`, `markdown`, `engine`, or an `error` with a
`code` and a `message`. The exit code is 1 when every URL failed.

`--public-only` is for fetching URLs you do not trust, such as ones a user
typed into a server. It is also on when the config sets `fetch.publicOnly:
true` or the environment sets `SONAR_FETCH_PUBLIC_ONLY=1`; neither can turn
it off. When it is on:

- sonar refuses non-http(s) schemes, URLs with credentials, `localhost`,
  single-label names, `.local`, `.internal`, `.localhost` (and similar)
  suffixes, and numeric hosts a browser would read as IPv4.
- sonar resolves the host and refuses it when ANY address is loopback,
  private, link-local, CGNAT, ULA, multicast, unspecified or reserved,
  including IPv4-mapped, NAT64 and 6to4 addresses that embed one.
- The HTTP path dials the address it checked, re-checks the connected
  address in the dialer's `Control` hook, and re-checks every redirect hop.
- Chromium runs behind a loopback proxy inside sonar. Every request it makes
  (navigations, 3xx hops, JavaScript and meta-refresh redirects, iframes,
  subresources, websockets) goes through that proxy, which checks the host
  and connects to the checked address. Chromium resolves no names itself
  (`--host-resolver-rules`), skips QUIC, and keeps WebRTC to proxied traffic.
- A page that triggers any refused request fails with no content.
- The Parallel extract API is skipped, since sonar cannot check the hops a
  third party follows.

Error codes: `blocked_url` (scheme, credentials or internal name),
`blocked_internal` (the host or a request reached an internal address),
`blocked_unresolved` (no address to check), `fetch_failed` (anything else).

Because the proxy and the HTTP dialer connect to the exact address they
checked, a DNS answer that changes between check and connect (rebinding) is
never used. The remaining exposure is a public address you do not consider
public, for example a host on your own public subnet: firewall the fetcher's
egress if that matters.

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

The tool is `search(query, count, engines)`: it runs the same dispatch, dedupe
and ranking as `sonar search`, so both surfaces return identical results for
the same query. `engines` is an optional list of engine names to run, so an
agent can target one engine (research code with Exa, generic web with Tavily)
or a subset; omitted, every enabled engine runs. Results are cached and
throttled exactly as on the CLI. API keys
are read from the environment the same way the CLI reads them, which an agent
launched from a normal shell inherits — no extra setup.

## Configuration

`~/.sonar.yml`, or the path in `SONAR_CONFIG`. A missing file is not an error —
the defaults stand. Top-level keys: `count`, `timeout`, `minInterval`,
`jitter`, `breakerCooldown`, `cacheDir`, `cacheTTL`, `searxng`,
`searxngPriority`, `fetch.publicOnly`, and `engines.*` with `enabled`, `apiKey`, `baseURL`,
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

## Go library

`github.com/FacileStudio/sonar/websearch` runs the same dispatcher from Go.
Keys come only from its `Config`: it reads no config file, no environment and
no tiroir, and caches results in memory.

```go
c, err := websearch.New(websearch.Config{Keys: map[string]string{"tavily": key}})
results, err := c.Search(ctx, "query", 5, nil)
```

`Client.Fetch` fetches one page as markdown, with the same guard as
`sonar fetch --public-only` turned on by default:

```go
page, err := c.Fetch(ctx, "https://example.com", websearch.FetchOptions{})
// page.Markdown, page.FinalURL, page.Status, page.Title, page.Engine, page.Truncated
var blocked *websearch.BlockedError
if errors.As(err, &blocked) {
	// blocked.Code: websearch.CodeBlockedURL, CodeBlockedInternal or CodeBlockedUnresolved
}
```

`FetchOptions`: `Timeout` (default 30s), `MaxBytes` (default 2 MiB, caps the
HTML read and the markdown; `Truncated` says when it cut), `Browser` (render in
headless Chromium; rod downloads one when none is installed), `Selector` (CSS
selector, needs `Browser`, else `ErrSelectorNeedsBrowser`), and
`AllowInternal`. Without `Browser`, Fetch uses plain HTTP and never starts a
browser. A non-2xx status is returned in `Page.Status`, not as an error.
`AllowInternal: true` turns the guard off: use it only for trusted callers
fetching trusted URLs, since a redirect can then reach loopback, private
networks and cloud metadata.

## Hosted service

`sonar serve` runs the hosted search engine at https://sonar.facile.studio: a web UI (web, images and news tabs, keyboard navigation, suggestions, `!bangs`, OpenSearch), a JSON API under `/api` with its reference at `/docs`, and an image proxy. People sign in with their Facile account. Queries are never logged or stored; each search on sonar's own keys counts once against a daily quota set by the person's plan, and searches on keys they bring themselves do not count. See [docs/hosted.md](docs/hosted.md).

## Exit codes

`0` success. `1` error — a failing search, or a missing query.