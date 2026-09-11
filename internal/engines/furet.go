package engines

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/FacileStudio/sonar/internal/httpc"
)

// Furet queries the self-hosted SearXNG instance. It aggregates many upstream
// engines in one call, so it is friendlier than scraping a single engine, but
// it inherits the upstream engines' IP blocks from ruche's own SearXNG config.
type Furet struct {
	BaseURL string
	Client  *http.Client
	Count   int
}

func (f *Furet) Name() string { return "furet" }

type searxResponse struct {
	Results []struct {
		Title   string `json:"title"`
		URL     string `json:"url"`
		Content string `json:"content"`
	} `json:"results"`
}

func (f *Furet) Search(ctx context.Context, query string, count int) ([]Result, error) {
	u := f.BaseURL + "/search?" + url.Values{
		"q":      {query},
		"format": {"json"},
	}.Encode()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	httpc.Fingerprint(req.Header)
	req.Header.Set("Accept", "application/json")
	res, err := f.Client.Do(req)
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
		out = append(out, Result{Title: r.Title, URL: r.URL, Snippet: r.Content, Engine: f.Name()})
	}
	return out, nil
}
