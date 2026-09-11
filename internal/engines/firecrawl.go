package engines

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/FacileStudio/sonar/internal/httpc"
)

// Firecrawl queries the Firecrawl Search API (a keyed, provider-metred web +
// scrape service), so ruche's IP reputation never applies. It combines search
// with content scraping and can return full markdown, but for sonar we only
// need the standard title/description/url results.
type Firecrawl struct {
	Key    string
	Client *http.Client
	Count  int
}

func (f *Firecrawl) Name() string { return "firecrawl" }

type firecrawlRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

type firecrawlResponse struct {
	Data struct {
		Web []struct {
			Title       string `json:"title"`
			URL         string `json:"url"`
			Description string `json:"description"`
		} `json:"web"`
	} `json:"data"`
}

func (f *Firecrawl) Search(ctx context.Context, query string, count int) ([]Result, error) {
	if f.Key == "" {
		return nil, ErrBlocked
	}
	body, _ := json.Marshal(firecrawlRequest{
		Query: query,
		Limit: countAt(count, f.Count),
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.firecrawl.dev/v2/search", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTransient, err)
	}
	httpc.Fingerprint(req.Header)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+f.Key)
	res, err := f.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTransient, err)
	}
	defer res.Body.Close()
	switch {
	case res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden:
		return nil, ErrBlocked
	case res.StatusCode == http.StatusTooManyRequests || res.StatusCode == http.StatusRequestTimeout || res.StatusCode >= 500:
		return nil, ErrTransient
	case res.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("%w: status %d", ErrTransient, res.StatusCode)
	}
	var parsed firecrawlResponse
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("%w: decode: %v", ErrTransient, err)
	}
	out := make([]Result, 0, len(parsed.Data.Web))
	for _, r := range parsed.Data.Web {
		out = append(out, Result{Title: r.Title, URL: r.URL, Snippet: r.Description, Engine: f.Name()})
	}
	return out, nil
}
