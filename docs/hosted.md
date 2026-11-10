# Hosted sonar (`sonar serve`)

`sonar serve` runs the hosted search engine at https://sonar.facile.studio: a web UI, a JSON API under `/api`, an image proxy at `/img` and an OpenSearch description at `/opensearch.xml`. People sign in with their Facile account (Registre). The API reference is at `/docs`.

## Privacy rules the code enforces

- No search history. The only per-person data is one counter per UTC day (`search_usage`), deleted after two days, plus the engine keys a person saved.
- No query in any log. The request logger redacts `q`, `u`, `s` and `url`; engine errors are never logged with their URL; the compose file turns Traefik access logs off for sonar's routers, because the request path carries the query.
- Brave results are never cached, in the server or the CLI (`engines.Uncacheable`). Other engines keep sonar's in-memory cache.
- Result images load through `/img`, which serves only URLs sonar signed, only from the public internet (netguard), only real images (sniffed, no SVG), capped at 5 MiB, with no cookies.
- External `!bangs` send the query to the named site; the UI says so on `/bangs`. Typing suggestions go to DuckDuckGo without the person's identity and can be turned off in settings or with `SONAR_SUGGEST_URL=off`.

## Engine policy

A web search on sonar's own keys walks a chain, one engine at a time: Brave first, then Tavily only when Brave returned fewer than 5 results, errored, or took longer than 6 seconds. A search therefore calls 1 engine, 2 at most. `SONAR_WEB_ENGINES` sets the chain (`all` restores the old run on every engine at once) and `SONAR_WEB_ENOUGH` the threshold. Engines in the chain without a key are skipped; an empty chain falls back to every engine. Image and news searches call Brave only, the one engine sonar has for them. A `!bang` that names an engine, and searches on a person's own keys, run exactly what they ask. The CLI and the `websearch` library keep their own defaults.

## Quotas

Every search, image search, news search and page fetch on sonar's own keys spends one search of the daily quota. The quota is the forfait entitlement `sonar.searches_daily` (a limit, `-1` unlimited) read from the Registre directory snapshot. Until Registre has served an entitlements snapshot, everyone gets the free default (20). An exhausted quota answers 402 `payment_required` with `upgrade_url`. A search no engine answers is refunded.

Searches on a person's own keys (Settings, Engine keys) skip the quota. Keys are sealed with porte's AES-256-GCM sealer under `PORTE_TOKEN_KEY`; without that key the server stores no user key.

Proposed plan values: free 20, Pro 500, Studio 2000. They are set in Ardoise (billing, feature `sonar.searches_daily`), never in sonar.

## AI summary

`POST /api/summary` takes only `{"query"}`. sonar finds the sources itself on its summary engines (Tavily, then Exa, then Parallel, never Brave) and asks Bulle to summarize them. It needs `ai.access`, counts as one search, and is off unless `SONAR_AI_SUMMARY=true`, `BULLE_URL` and `SONAR_BULLE_KEY` are set. See [ai-summary.md](ai-summary.md).

## MCP

`/mcp` is reserved at the root, next to `/api`, and will reuse the same search service. Owner decision (2026-11-08): MCP access is free for every account, third-party agents included, so `/mcp` must never be plan-gated. Searches made through it still spend the search quota.

## Settings

| Variable | Meaning |
|---|---|
| `DATABASE_URL` | PostgreSQL. Required. |
| `PORT` | Listen port, 4000 in the image. |
| `SONAR_PUBLIC_URL` | The public origin, used by OpenSearch and the OIDC defaults. |
| `OIDC_ISSUER`, `OIDC_CLIENT_ID`, `OIDC_CLIENT_SECRET` | The Registre client. Without an issuer nobody can sign in. |
| `OIDC_REDIRECT_URL`, `OIDC_SUCCESS_URL` | Default to `<public>/api/auth/oidc/callback` and `<public>/`. |
| `PORTE_TOKEN_KEY` | 32-byte hex or base64 key: seals porte tokens and user engine keys. |
| `SONAR_<ENGINE>_KEY` | sonar's own keys: brave, tavily, exa, parallel, linkup, firecrawl, serpapi. Images and news need Brave. |
| `SONAR_SEARXNG` | Comma-separated SearXNG instances. |
| `SONAR_SUGGEST_URL` | Suggestion endpoint (OpenSearch format). Default DuckDuckGo; `off` disables. |
| `SONAR_IMAGE_KEY` | Pins the image-signing key (32+ bytes); random per boot otherwise. |
| `SONAR_WEB_ENGINES` | The web chain, default `brave,tavily`; `all` runs every engine at once. |
| `SONAR_WEB_ENOUGH` | Results that end a chain early, default 5. |
| `SONAR_SUMMARY_ENGINES` | Where summary sources come from, default `tavily,exa,parallel`. `brave` is always dropped. |
| `SONAR_AI_SUMMARY`, `BULLE_URL`, `SONAR_BULLE_KEY` | The AI summary, off by default. |
| `TRUSTED_PROXIES` | `private,cloudflare` behind Traefik and Cloudflare. |

Registre calls `POST /api/registre/sync` (directory nudge, carries entitlements) and `POST /api/registre/deletions` (GDPR erasure).

## Development

```sh
sh scripts/check.sh                       # Go, client, filet
TEST_DATABASE_URL=postgres://... go test ./...   # with the database tests
cd apps/client && bun install && bun run build
CLIENT_DIR=apps/client/build DATABASE_URL=... go run . serve
```
