package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/FacileStudio/sonar/internal/engines"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type fetchInput struct {
	URL       string `json:"url" jsonschema:"the web address (http:// or https://) to fetch"`
	WaitUntil string `json:"wait_until,omitempty" jsonschema:"wait condition: networkidle, domcontentloaded, load, or none (default: networkidle)"`
	Selector  string `json:"selector,omitempty" jsonschema:"optional CSS selector to restrict extraction to a specific container"`
}

type fetchOutput struct {
	Title   string `json:"title" jsonschema:"the page title"`
	URL     string `json:"url" jsonschema:"the fetched page URL"`
	Content string `json:"content" jsonschema:"the extracted markdown content"`
}

func runFetch(ctx context.Context, _ *mcp.CallToolRequest, in fetchInput) (*mcp.CallToolResult, fetchOutput, error) {
	url := strings.TrimSpace(in.URL)
	if url == "" {
		return nil, fetchOutput{}, errors.New("fetch needs a URL")
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return nil, fetchOutput{}, errors.New("URL must start with http:// or https://")
	}

	rodEngine := &engines.Rod{
		Count:         1,
		WaitCondition: in.WaitUntil,
		Selector:      in.Selector,
	}
	results, err := rodEngine.Search(ctx, url, 1)
	if err != nil {
		return nil, fetchOutput{}, err
	}
	if len(results) == 0 {
		return nil, fetchOutput{}, fmt.Errorf("no content extracted from %s", url)
	}

	out := fetchOutput{
		Title:   results[0].Title,
		URL:     results[0].URL,
		Content: results[0].Snippet,
	}
	return nil, out, nil
}
