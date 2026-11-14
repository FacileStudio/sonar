# AI summary: what sonar expects from Bulle

sonar holds no AI key. Its summary calls Bulle, the suite's AI hub, the same way other services call `POST /api/ai/decide`: a service key plus the person it acts for. `/api/ai/decide` only answers choice questions, so Bulle serves a text route for sonar, `POST /api/ai/summarize`. sonar calls it through tronc's `bulle` client, which never follows a redirect.

## Sources never come from Brave

Owner decision (2026-11-08): the summary never uses Brave results. The client sends only the query (`POST /api/summary {"query": "…"}`). sonar then runs its own search on the summary chain, `SONAR_SUMMARY_ENGINES` (default `tavily,exa,parallel`, one engine at a time until 5 results are in), and sends those sources to Bulle. `brave` is removed from that list whatever the setting says, and the Brave results on the page are never sent. The answer returns the sources it cites, so the client lists them under the summary. A summary costs one search of the daily quota, refunded if no engine answers.

## Route

```
POST /api/ai/summarize
Authorization: Bearer <sonar's service key>        (listed in BULLE_AI_CALLERS as sonar:<sha256>)
X-Bulle-Acting-For: <the person's email>
Content-Type: application/json

{"query": "…", "sources": [{"title": "…", "url": "https://…", "snippet": "…"}]}   (1 to 10 sources)

200 {"summary": "… [1] … [2]", "model": "…"}
402 payment_required (the acting person lacks ai.access)
```

Bulle refuses a source whose title is over 300 characters or whose URL is over 2048 bytes. sonar cuts titles to 300 characters and snippets to 600, and leaves out a source whose URL is too long, so one tracking link never fails the whole summary. sonar also keeps the whole body under 30 KiB, below Bulle's 32 KiB cap, by dropping sources from the end; the sources it keeps stay in order, so `[n]` still points at the right one.

sonar answers the person with Bulle's 402 and its upgrade link, Bulle's 429 as "too many summaries; try again in a minute" (the route has no daily cap; Bulle's per-caller and per-user burst limiter answers 429 in plain text, and tronc's client drops its `Retry-After`), any other refusal as a 500 (a sonar bug or a wrong service key, not worth retrying), and only a 5xx, timeout or network error as 503 "try again".

Bulle should check `ai.access` for the acting person (its `Gate.CheckAI`), cap the body (32 KiB), rate limit per caller, log usage without the query, and cite sources by their 1-based position. sonar checks `ai.access` itself first, so a free account never reaches Bulle.

## Turning it on

Add sonar's key hash to Bulle's `BULLE_AI_CALLERS`, set `BULLE_URL`, `SONAR_BULLE_KEY` and `SONAR_AI_SUMMARY=true` on the sonar compose, and check that at least one summary engine has a key (`summary_engines` in the boot log).
