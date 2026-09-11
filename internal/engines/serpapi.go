package engines

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/FacileStudio/sonar/internal/httpc"
)

// SerpAPI queries the SerpAPI service (a keyed provider that runs search
// engines like Google from its own servers), so ruche's IP reputation never
// applies. Returns the provider's organic results.
type SerpAPI struct {
	Key    string
	Client *http.Client
	Count  int
}

func (s *SerpAPI) Name() string { return "serpapi" }

type serpapiResponse struct {
	OrganicResults []struct {
		Title   string `json:"title"`
		Link    string `json:"link"`
		Snippet string `json:"snippet"`
	} `json:"organic_results"`
}

func (s *SerpAPI) Search(ctx context.Context, query string, count int) ([]Result, error) {
	if s.Key == "" {
		return nil, ErrBlocked
	}
	u := "https://serpapi.com/search?" + url.Values{
		"engine":  {"google"},
		"q":       {query},
		"api_key": {s.Key},
		"num":     {fmt.Sprint(countAt(count, s.Count))},
	}.Encode()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	httpc.Fingerprint(req.Header)
	req.Header.Set("Accept", "application/json")
	res, err := s.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTransient, err)
	}
	defer res.Body.Close()
	switch {
	case res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden:
		return nil, ErrBlocked
	case res.StatusCode == http.StatusTooManyRequests || res.StatusCode >= 500:
		return nil, ErrTransient
	case res.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("%w: status %d", ErrTransient, res.StatusCode)
	}
	var parsed serpapiResponse
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("%w: decode: %v", ErrTransient, err)
	}
	out := make([]Result, 0, len(parsed.OrganicResults))
	for _, r := range parsed.OrganicResults {
		out = append(out, Result{Title: r.Title, URL: r.Link, Snippet: r.Snippet, Engine: s.Name()})
	}
	return out, nil
}
