package engines

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/FacileStudio/sonar/internal/httpc"
)

// Brave queries the Brave Search API from the provider's servers, so ruche's
// IP reputation never applies. Never rate-throttled client-side: Brave is
// metered by its own monthly quota.
type Brave struct {
	Key    string
	Client *http.Client
	Count  int
}

func (b *Brave) Name() string { return "brave" }

type braveResponse struct {
	Web struct {
		Results []struct {
			Title       string `json:"title"`
			URL         string `json:"url"`
			Description string `json:"description"`
		} `json:"results"`
	} `json:"web"`
}

func (b *Brave) Search(ctx context.Context, query string, count int) ([]Result, error) {
	if b.Key == "" {
		return nil, ErrBlocked
	}
	u := "https://api.search.brave.com/res/v1/web/search?" + url.Values{
		"q":     {query},
		"count": {fmt.Sprint(countAt(count, b.Count))},
	}.Encode()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	httpc.Fingerprint(req.Header)
	req.Header.Set("X-Subscription-Token", b.Key)
	req.Header.Set("Accept", "application/json")
	var parsed braveResponse
	if derr := decode(b.Client, req, &parsed, 0); derr != nil {
		return nil, derr
	}
	out := make([]Result, 0, len(parsed.Web.Results))
	for _, r := range parsed.Web.Results {
		out = append(out, Result{Title: r.Title, URL: r.URL, Snippet: r.Description, Engine: b.Name()})
	}
	return out, nil
}
