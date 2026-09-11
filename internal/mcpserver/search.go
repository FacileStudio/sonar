package mcpserver

import (
	"context"
	"errors"
	"strings"

	"github.com/FacileStudio/sonar/internal/config"
	"github.com/FacileStudio/sonar/internal/engines"
	"github.com/FacileStudio/sonar/internal/search"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// searchInput is what a model may ask for. Query is the only required field.
type searchInput struct {
	Query string `json:"query" jsonschema:"the words to search for across the enabled engines"`
	Count int    `json:"count,omitempty" jsonschema:"max results to return; 0 means the configured default"`
}

// searchHit is one ranked result, carrying enough to open the page it came from.
type searchHit struct {
	Title   string `json:"title" jsonschema:"result title"`
	URL     string `json:"url" jsonschema:"the result's web address"`
	Snippet string `json:"snippet,omitempty" jsonschema:"a sample line from the page"`
	Engine  string `json:"engine" jsonschema:"which engine returned the result"`
}

// searchOutput is the ranked, deduped result list.
type searchOutput struct {
	Results []searchHit `json:"results"`
}

// runSearch runs the query through the same ranked search the CLI uses, so the
// MCP surface and `sonar search` agree result-for-result on the same query.
func runSearch(ctx context.Context, _ *mcp.CallToolRequest, in searchInput) (*mcp.CallToolResult, searchOutput, error) {
	query := strings.TrimSpace(in.Query)
	if query == "" {
		return nil, searchOutput{}, errors.New("search needs a query")
	}
	cfg, err := config.Load(config.Path())
	if err != nil {
		return nil, searchOutput{}, err
	}
	count := in.Count
	if count <= 0 {
		count = cfg.DefaultCount
	}
	results, err := search.Query(ctx, cfg, query, count)
	if err != nil {
		return nil, searchOutput{}, err
	}
	return nil, searchOutput{Results: hits(results)}, nil
}

// hits converts engine results into the declared output shape.
func hits(in []engines.Result) []searchHit {
	out := make([]searchHit, 0, len(in))
	for _, r := range in {
		out = append(out, searchHit{Title: r.Title, URL: r.URL, Snippet: r.Snippet, Engine: r.Engine})
	}
	return out
}
