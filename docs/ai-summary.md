# AI summary: what sonar expects from Bulle

sonar holds no AI key. Its summary calls Bulle, the suite's AI hub, the same way other services call `POST /api/ai/decide`: a service key plus the person it acts for. `/api/ai/decide` only answers choice questions, so sonar needs a text route. Until Bulle ships it, `SONAR_AI_SUMMARY` stays `false`.

## Proposed route

```
POST /api/ai/summarize
Authorization: Bearer <sonar's service key>        (listed in BULLE_AI_CALLERS as sonar:<sha256>)
X-Bulle-Acting-For: <the person's email>
Content-Type: application/json

{"query": "…", "sources": [{"title": "…", "url": "https://…", "snippet": "…"}]}   (1 to 10 sources)

200 {"summary": "… [1] … [2]", "model": "…"}
402 payment_required (the acting person lacks ai.access)
```

Bulle should check `ai.access` for the acting person (its `Gate.CheckAI`), cap the body (32 KiB), rate limit per caller, log usage without the query, and cite sources by their 1-based position. sonar checks `ai.access` itself first, so a free account never reaches Bulle.

## Open legal point

Brave's API terms forbid derivative works of search results. Whether an AI summary of Brave results is one is for the owner to decide before turning the summary on; restricting the summary to non-Brave results is the fallback.
