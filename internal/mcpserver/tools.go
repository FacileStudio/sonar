package mcpserver

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// searchTool declares the web search.
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

// fetchTool declares the webpage fetch tool.
func fetchTool() *mcp.Tool {
	return &mcp.Tool{
		Name:  "fetch",
		Title: "Fetch webpage content",
		Description: "Fetch a webpage URL and extract its main content as clean Markdown using a stealth headless browser. " +
			"Renders client-side JavaScript, handles dynamic single-page applications, and strips noise (scripts, styles, ads). " +
			"Returns the page title, canonical URL, and markdown text.",
		Annotations: &mcp.ToolAnnotations{
			ReadOnlyHint:    true,
			DestructiveHint: new(false),
			IdempotentHint:  true,
			OpenWorldHint:   new(true),
		},
	}
}
