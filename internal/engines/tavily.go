package engines

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/FacileStudio/sonar/internal/httpc"
)

// Tavily queries the Tavily Search API (a keyed, provider-metred service, so
// ruche's IP reputation never applies). Unlike the free search APIs it is
// designed for LLM/agent use, returning clean tokens and URLs.
type Tavily struct {
	Key    string
	Client *http.Client
	Count  int
}

func (t *Tavily) Name() string { return "tavily" }

type tavilyRequest struct {
	Query         string `json:"query"`
	MaxResults    int    `json:"max_results"`
	SearchDepth   string `json:"search_depth,omitempty"`
	IncludeAnswer bool   `json:"include_answer,omitempty"`
}

type tavilyResponse struct {
	Results []struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Content string `json:"content"`
	} `json:"results"`
}

func (t *Tavily) Search(ctx context.Context, query string, count int) ([]Result, error) {
	if t.Key == "" {
		return nil, ErrBlocked
	}
	body, _ := json.Marshal(tavilyRequest{
		Query:      query,
		MaxResults: countAt(count, t.Count),
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.tavily.com/search", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTransient, err)
	}
	httpc.Fingerprint(req.Header)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.Key)
	var parsed tavilyResponse
	if derr := decode(t.Client, req, &parsed, 0); derr != nil {
		return nil, derr
	}
	out := make([]Result, 0, len(parsed.Results))
	for _, r := range parsed.Results {
		out = append(out, Result{Title: r.Title, URL: r.URL, Snippet: r.Content, Engine: t.Name()})
	}
	return out, nil
}
