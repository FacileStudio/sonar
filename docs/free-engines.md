# Free engines for the hosted chain

Goal: spend Brave and Tavily keys less by answering more searches from free engines first.

## What the hosted chain can use today

The web chain walks `SONAR_WEB_ENGINES` in order, one engine at a time. It stops once `SONAR_WEB_ENOUGH` (default 5) distinct results are in. An engine that errors, times out (3 seconds for a free engine, 6 for a keyed one, except the last one) or returns too few results passes the turn to the next one. So an order such as `searxng,brave,tavily` already falls back correctly. This is covered by `TestWebChainTriesFreeEnginesBeforePaidOnes`.

One free engine is reachable by env alone:

```
SONAR_SEARXNG=https://searx.internal.example
SONAR_WEB_ENGINES=searxng,brave,tavily
```

That needs a SearXNG instance with `json` listed under `search.formats`. With several URLs in `SONAR_SEARXNG` the engines are named `searxng1`, `searxng2` and so on, and each goes into the chain by that name.

DuckDuckGo (`ddg`) and Bing (`bing`) exist as scrapers in `internal/engines/` but the hosted server could not switch them on: `websearch.Config` only enabled keyed engines and SearXNG. Both scrape from a datacenter IP and are often blocked (DDG answers 202 or 429), so they are best-effort. Their breaker parks them for two minutes after a block.

Engines that need a key and cost money or have a quota: brave, tavily, exa, parallel, linkup, firecrawl, serpapi (Brave and Tavily are the web chain default).

## What production runs

One SearXNG container, `sonar-searxng`, in the same compose as `sonar-api`:

```
SONAR_SEARXNG=http://sonar-searxng:8080
SONAR_FREE_ENGINES=ddg
SONAR_WEB_ENGINES=searxng,ddg,brave,tavily
SEARXNG_SECRET=<random, read by the SearXNG container>
```

- Image: `deploy/searxng/Dockerfile` pins an official dated tag (SearXNG is a rolling release with no version numbers; tags are listed at https://hub.docker.com/r/searxng/searxng/tags) and bakes `deploy/searxng/settings.yml` in. To change a setting, edit that file and deploy.
- Reach: the compose network only. No published port, no Traefik label. The limiter stays off, since nothing but `sonar-api` can call it.
- Settings: `search.formats` is `json` only (https://docs.searxng.org/admin/settings/settings_search.html), the secret key comes from `$SEARXNG_SECRET` and the limiter is off (https://docs.searxng.org/admin/settings/settings_server.html), `general.debug` is false (https://docs.searxng.org/admin/settings/settings_general.html).
- Logs: the container has none (`logging: driver: none`). SearXNG has no setting that keeps queries out of its logs. `debug: false` only stops the verbose logging; at its production level (WARNING) it still prints the full upstream URL when an engine request fails, for example `WARNING:searx.network.mojeek: HTTP Request failed: GET https://www.mojeek.com/search?q=...` (seen on 2026-11-15 with the pinned image). The host's log collectors tail every container, so the only safe level is no log at all. The Granian access log is off by default and pinned off (`GRANIAN_LOG_ACCESS_ENABLED=false`). Sonar posts the query in a form body, so it is in no URL on the way in either.
- Engines: mojeek, startpage and yahoo, the key-less engines that answered from the server's address on 2026-11-15. Google and Qwant denied it, DuckDuckGo answered with a captcha, Bing returned nothing. SearXNG's Brave scraper is left out: Sonar is a Brave API customer.
- Time: SearXNG gives its engines 2 seconds (2.5 at most), and the chain gives a free engine 3 seconds before moving on, so a search reaches Brave after 6 seconds at worst when both free engines hang. A blocked `ddg` is parked for two minutes and costs nothing while parked.
- Debugging: run the image by hand with a throwaway `SEARXNG_SECRET` and read its output there. `docker inspect` still shows the health status of the production container.

## What the pool and the scraper switch add (all off by default)

| Env | Effect |
|---|---|
| `SONAR_SEARXNG_POOL` | Comma list of SearXNG base URLs. One engine named `searxpool` rotates through them. |
| `SONAR_SEARXNG_POOL_TRIES` | Instances one search may try. Default 2. |
| `SONAR_FREE_ENGINES` | `ddg`, `bing` or both: switches the key-less scrapers on. |

Pool behaviour: start at the next instance in rotation, at most N tries per search, 3 second timeout per attempt, a circuit breaker per instance (the existing `politeness.Breaker`, cooldown from `breakerCooldown`, two minutes by default). A 403, a 429, an HTML page where JSON was expected (a captcha or a login wall) and a JSON decode failure all park that instance. The pool never retries a refusal and never tries to get past a bot wall. It sends an honest `User-Agent` (`sonar-search (+https://sonar.facile.studio)`), not a browser fingerprint. A pool that is fully parked returns a transient error, so the chain moves on to the next engine at once.

Example chain with free engines first:

```
SONAR_WEB_ENGINES=searxpool,ddg,brave,tavily
SONAR_SEARXNG_POOL=https://searx.internal.example
SONAR_FREE_ENGINES=ddg
```

No public instance is listed by default. Each URL in the pool is a choice the owner makes after reading that operator's rules.

## Privacy

A public instance receives every query sent to it, and Sonar cannot see or control what its operator logs, while Sonar itself still logs and stores no query. The query goes in a POST body, never in a URL, and errors name neither the query nor a URL; `TestPoolNeverShowsTheQueryOrTheRequestURL` checks both, plus the pool's breaker state. There are no per-engine metrics in the repo, so none carries a query.

## Research: public SearXNG

Read and cited:

- The official Search API page says `format=json` works only when the instance lists it in `settings.yml`, that an unset format returns 403 Forbidden, and that "many public instances have these formats disabled". https://docs.searxng.org/dev/search_api.html
- `search.formats` is the setting that controls this. https://docs.searxng.org/admin/settings/settings_search.html
- A SearXNG maintainer (return42) wrote in a bug report: "In general, public instances do not support JSON, primarily to protect themselves against bots. If specific instances have enabled JSON access, they employ their own individual bot protection measures. Consequently, any attempt to retrieve JSON data from a public instance is bound to fail." https://github.com/searxng/searxng/issues/6676
- The public instance list is a directory of instances with uptime and response times. https://searx.space/ . I found no rule about bot or API use on that page. The instance listing rules are at https://github.com/searxng/searx-instances and also say nothing about it. Whether each instance allows API use is up to its own operator.
- A third-party guide (not official) advises reading each instance's policy first, making one JSON request, stopping after a denial, never bypassing a captcha, and not sending a query to several independent operators needlessly. https://github.com/B0TTS/searxng-mcp/blob/main/docs/public-searxng-instances.md

Verdict. Allowed: only by an instance that has enabled JSON and whose own policy permits automation. Discouraged by the SearXNG project: using public instances as an API. Forbidden: not stated globally; per instance, unverified.

Smoke test, 2 requests from this server's IP, harmless query, status codes only: `paulgo.io` answered 429 (text/html); `baresearch.org` answered 200 with `text/html` (no JSON). Neither gave results. Both are listed on searx.space; I did not read their policies, so this says nothing about whether they allow bots. Result: public instances are not a dependable source today.

## Research: 4get

- 4get documents a JSON API (`/api/v1/web`, `/images`, `/videos`, `/news`, `/music`) and a status field (`"ok"`). https://git.lolcat.ca/lolcat/4get/raw/branch/master/api.txt and https://4get.nadeko.net/api-docs
- Terms of use on both: do not do "Serp SEO scanning", "Intensive scraping", or "Any other activity that isn't triggered by a human"; cache responses; good uses include "A chatroom bot that presents users with search results" and personal use. Operators with a pass system require a captcha-derived pass cookie, worth about 100 requests per 24 hours. The same docs say a bad pass returns HTTP 429.
- `https://4get.ca/ami4get` reports `bot_protection: 1` and `api_enabled: true` for 4get.ca, and lists 22 other instances. https://4get.ca/ami4get
- The docs do not specify the shape of a web result. They call the rest of the JSON "pretty self explanatory". Only `answer`, `spelling`, `npt` and the image shape are written down.

Verdict: not implemented. The result array and its fields are undocumented, I did not guess, and a pass-gated captcha cannot be solved by a server. A search proxy for many users also sits badly with "not triggered by a human" and "intensive" limits. Revisit if an operator gives written permission and a stable shape.

## Research: other free engines

- Marginalia: free key by email, licensed CC-BY-NC-SA 4.0 (non-commercial). The shared key `public` "often hits a rate limit". Commercial use is a paid metered key. A small-site index, not general web. Not added. https://about.marginalia-search.com/article/api/
- Mojeek: paid pay-as-you-go, from 2 pounds per 1,000 requests, with a free credit but a bank card is required. Not free. https://www.mojeek.com/services/search/web-search-api/
- Wikipedia (MediaWiki action API): free, asks for an informative User-Agent and serial requests. Only searches Wikipedia, so it suits a vertical, not the web chain. Not added. https://www.mediawiki.org/wiki/API:Etiquette
- Brave free tier is already the first engine in the chain.

## Unverified

- Each public instance's own bot policy (none read).
- Whether JSON is enabled on any listed instance beyond the two probed.
- 4get web result field names.
- The Wikimedia rate limits page was only seen in a search snippet: https://www.mediawiki.org/wiki/Wikimedia_APIs/Rate_limits
