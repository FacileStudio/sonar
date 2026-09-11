# sonar

Multi-engine web search for the terminal. sonar runs one query across keyed
search APIs (Brave, Tavily, Exa, Firecrawl, SerpApi), your SearXNG instance,
and scrape fallbacks (Bing, DuckDuckGo), then merges the results into a single
ranked list. It is built to dodge IP-based blocking: engines are dispatched in
priority order with throttling and jitter, a per-engine circuit breaker parks
any backend that starts failing, and responses are cached so a repeated query
is not refetched.

## Engines

| Engine | Kind | Default |
| --- | --- | --- |
| furet | SearXNG instance (https://furet.facile.studio) | on |
| brave | keyed API | off |
| tavily | keyed API | off |
| exa | keyed API | off |
| firecrawl | keyed API | off |
| serpapi | keyed API | off |
| bing | scrape | off |
| ddg | scrape | off |

## Usage

```sh
sonar search "circuit breaker pattern"
sonar search -n 5 "go 1.26 modules"
sonar search --json "lipgloss colors"
```

`-n, --count N` caps the number of results (default: the config `count`, 10).
`--json` prints one JSON document on stdout and nothing else.

## Configuration

`~/.config/sonar/config.yml`, or the path in `SONAR_CONFIG`. A missing file is
not an error — the defaults stand. Top-level keys: `count`, `timeout`,
`minInterval`, `jitter`, `breakerCooldown`, `cacheDir`, `cacheTTL`, and
`engines.*` with `enabled`, `apiKey`, `baseURL`, `priority`, `count` per
engine.

API keys belong in the file or in the environment: `SONAR_<ENGINE>_KEY`
(`SONAR_BRAVE_KEY`, `SONAR_TAVILY_KEY`, ...). The common provider names are
read as aliases: `EXA_KEY`, `FIRECRAWL_KEY`, `TAVILY_KEY`, `SERPAPI_API_KEY`.
The environment is only consulted when the file has no key for that engine.

## Exit codes

`0` success. `1` error — a failing search, or a missing query.