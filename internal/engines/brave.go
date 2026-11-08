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
	Key     string
	Client  *http.Client
	Count   int
	BaseURL string
}

func (b *Brave) base() string {
	if b.BaseURL != "" {
		return b.BaseURL
	}
	return "https://api.search.brave.com"
}

func (b *Brave) Name() string { return "brave" }

// Uncached reports that Brave results must never be stored: Brave's Search API
// terms forbid caching results beyond transient operation.
func (b *Brave) Uncached() bool { return true }

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
	u := b.base() + "/res/v1/web/search?" + url.Values{
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
		out = append(out, Result{Title: Plain(r.Title), URL: r.URL, Snippet: Plain(r.Description), Engine: b.Name()})
	}
	return out, nil
}
