package engines

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/FacileStudio/sonar/internal/httpc"
)

// Parallel queries the Parallel.ai Search and Extract APIs for fast, LLM-optimized web results.
type Parallel struct {
	Key    string
	Client *http.Client
	Count  int
}

// Name returns the engine identifier.
func (p *Parallel) Name() string { return "parallel" }

type parallelAdvancedSettings struct {
	MaxResults  int  `json:"max_results,omitempty"`
	FullContent bool `json:"full_content,omitempty"`
}

type parallelSearchRequest struct {
	SearchQueries    []string                 `json:"search_queries"`
	Mode             string                   `json:"mode,omitempty"`
	AdvancedSettings parallelAdvancedSettings `json:"advanced_settings"`
}

type parallelSearchResponse struct {
	Results []struct {
		URL      string   `json:"url"`
		Title    string   `json:"title"`
		Excerpts []string `json:"excerpts"`
	} `json:"results"`
}

type parallelExtractRequest struct {
	URLs             []string                 `json:"urls"`
	AdvancedSettings parallelAdvancedSettings `json:"advanced_settings"`
}

type parallelExtractResponse struct {
	Results []struct {
		URL         string   `json:"url"`
		Title       string   `json:"title"`
		Excerpts    []string `json:"excerpts"`
		FullContent string   `json:"full_content"`
	} `json:"results"`
}

// Search runs a web search query through the Parallel.ai Search API.
func (p *Parallel) Search(ctx context.Context, query string, count int) ([]Result, error) {
	if p.Key == "" {
		return nil, ErrBlocked
	}
	body, _ := json.Marshal(parallelSearchRequest{
		SearchQueries:    []string{query},
		Mode:             "fast",
		AdvancedSettings: parallelAdvancedSettings{MaxResults: countAt(count, p.Count)},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.parallel.ai/v1/search", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTransient, err)
	}
	httpc.Fingerprint(req.Header)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.Key)
	var parsed parallelSearchResponse
	if derr := decode(p.Client, req, &parsed, 0); derr != nil {
		return nil, derr
	}
	out := make([]Result, 0, len(parsed.Results))
	for _, r := range parsed.Results {
		out = append(out, Result{
			Title:   r.Title,
			URL:     r.URL,
			Snippet: strings.Join(r.Excerpts, "\n\n"),
			Engine:  p.Name(),
		})
	}
	return out, nil
}

// Extract extracts web page content as markdown from one or more URLs using the Parallel.ai Extract API.
func (p *Parallel) Extract(ctx context.Context, urls []string) ([]Result, error) {
	if p.Key == "" {
		return nil, ErrBlocked
	}
	body, _ := json.Marshal(parallelExtractRequest{
		URLs:             urls,
		AdvancedSettings: parallelAdvancedSettings{FullContent: true},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.parallel.ai/v1/extract", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTransient, err)
	}
	httpc.Fingerprint(req.Header)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.Key)
	var parsed parallelExtractResponse
	if derr := decode(p.Client, req, &parsed, 0); derr != nil {
		return nil, derr
	}
	return parseExtractResults(&parsed, p.Name()), nil
}

func parseExtractResults(parsed *parallelExtractResponse, engine string) []Result {
	out := make([]Result, 0, len(parsed.Results))
	for _, r := range parsed.Results {
		snippet := r.FullContent
		if snippet == "" {
			snippet = strings.Join(r.Excerpts, "\n\n")
		}
		out = append(out, Result{
			Title:   r.Title,
			URL:     r.URL,
			Snippet: snippet,
			Engine:  engine,
		})
	}
	return out
}
