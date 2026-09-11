package engines

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/FacileStudio/sonar/internal/httpc"
)

// Searxng queries a SearXNG instance. It aggregates many upstream engines in
// one call, so it is friendlier than scraping a single engine, but it inherits
// the upstream engines' IP blocks from the instance's own SearXNG config. Label
// distinguishes multiple instances for the cache and merge layers; BaseURL must
// be set by the caller.
type Searxng struct {
	Label   string
	BaseURL string
	Client  *http.Client
	Count   int
}

func (s *Searxng) Name() string { return s.Label }

type searxResponse struct {
	Results []struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Content string `json:"content"`
	} `json:"results"`
}

func (s *Searxng) Search(ctx context.Context, query string, count int) ([]Result, error) {
	u := s.BaseURL + "/search?" + url.Values{
		"q":      {query},
		"format": {"json"},
	}.Encode()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	httpc.Fingerprint(req.Header)
	req.Header.Set("Accept", "application/json")
	res, err := s.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTransient, err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusTooManyRequests || res.StatusCode >= 500 {
		return nil, ErrTransient
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d", ErrTransient, res.StatusCode)
	}
	var parsed searxResponse
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("%w: decode: %v", ErrTransient, err)
	}
	out := make([]Result, 0, len(parsed.Results))
	for _, r := range parsed.Results {
		out = append(out, Result{Title: r.Title, URL: r.URL, Snippet: r.Content, Engine: s.Name()})
	}
	return out, nil
}
