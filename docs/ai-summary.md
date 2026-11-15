# AI summary: what sonar expects from Bulle

sonar holds no AI key. Its summary calls Bulle, the suite's AI hub, as the signed-in person: sonar exchanges the person's Registre session for a Bulle token (`token_exchange`, audience `bulle`, scopes `bulle.read` and `bulle.write`, through porte's `exchange`) and calls `POST /api/ai/summarize` with it. sonar holds no service key, so a leak cannot spend another person's AI. sonar calls it through tronc's `bulle` client (`Summarize`), which never follows a redirect.

## Sources never come from Brave

Owner decision (2026-11-08): the summary never uses Brave results. The client sends only the query (`POST /api/summary {"query": "…"}`). sonar then runs its own search on the summary chain, `SONAR_SUMMARY_ENGINES` (default `tavily,exa,parallel`, one engine at a time until 5 results are in), and sends those sources to Bulle. `brave` is removed from that list whatever the setting says, and the Brave results on the page are never sent. The answer returns the sources it cites, so the client lists them under the summary. A summary costs one search of the daily quota, refunded if no engine answers.

## Route

```
POST /api/ai/summarize
Authorization: Bearer <the person's exchanged token>
Content-Type: application/json

{"query": "…", "sources": [{"title": "…", "url": "https://…", "snippet": "…"}]}   (1 to 10 sources)

200 {"summary": "… [1] … [2]", "model": "…"}
402 payment_required (the acting person lacks ai.access)
```

Bulle refuses a source whose title is over 300 characters or whose URL is over 2048 bytes. sonar cuts titles to 300 characters and snippets to 600, and leaves out a source whose URL is too long, so one tracking link never fails the whole summary. sonar also keeps the whole body under 30 KiB, below Bulle's 32 KiB cap, by dropping sources from the end; the sources it keeps stay in order, so `[n]` still points at the right one.

sonar answers the person with Bulle's 402 and its upgrade link, Bulle's 429 with Bulle's own message when it has one (the monthly AI allowance is used up, with the reset date), else "too many summaries; try again in a minute" (the burst limiter answers in plain text), a person whose session cannot be exchanged (`exchange.ErrReconnect`) as "Sign in again to use AI summaries.", any other refusal as a 500 (a sonar bug, not worth retrying), and only a 5xx, timeout or network error as 503 "try again".

Bulle should check `ai.access` for the token's person (its `Gate.CheckAI`), cap the body (32 KiB), rate limit per user, log usage without the query, and cite sources by their 1-based position. sonar checks `ai.access` itself first, so a free account never reaches Bulle.

## Turning it on

Set `BULLE_URL` and `SONAR_AI_SUMMARY=true` on the sonar compose, with OIDC configured (Registre must grant sonar's client `token_exchange` to audience `bulle`), and check that at least one summary engine has a key (`summary_engines` in the boot log).
