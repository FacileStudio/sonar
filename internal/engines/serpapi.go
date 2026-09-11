package engines

import (
	"context"
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
	var parsed serpapiResponse
	if derr := decode(s.Client, req, &parsed, 0); derr != nil {
		return nil, derr
	}
	out := make([]Result, 0, len(parsed.OrganicResults))
	for _, r := range parsed.OrganicResults {
		out = append(out, Result{Title: r.Title, URL: r.Link, Snippet: r.Snippet, Engine: s.Name()})
	}
	return out, nil
}
