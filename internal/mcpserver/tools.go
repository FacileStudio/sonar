package mcpserver

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// searchTool declares the web search.
//
// ReadOnlyHint puts it in the client's read-only group, which a user approves
// once in bulk instead of confirming on every call. OpenWorldHint is true:
// unlike a memory search that reaches one configured host, sonar queries the
// open internet through several engines, so an agent should be told the world
// it touches is the public web.
func searchTool() *mcp.Tool {
	return &mcp.Tool{
		Name:  "search",
		Title: "Search the web",
		Description: "Run one query across the selected sonar engines — every enabled one by default, or the " +
			"subset named in 'engines' for a targeted search — merge the survivors, dedupe " +
			"by URL and rank by engine trust, then return one ranked result list. Each result " +
			"carries a title, URL, snippet and the engine that found it; results from scrape " +
			"fallbacks (bing, ddg) rank below keyed-API and SearXNG results, so trust the ordering " +
			"over any single position. Repeating the same query is served from cache.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    true,
			DestructiveHint: new(false),
			IdempotentHint:  true,
			OpenWorldHint:   new(true),
		},
	}
}
