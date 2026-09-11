package engines

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/FacileStudio/sonar/internal/httpc"
)

// Browserbase queries the Browserbase Search API — a keyed, provider-metred
// service whose search runs on their infrastructure, so ruche's IP reputation
// never applies. Returns token-efficient results meant to feed agent/RAG work.
// Note the responses carry no snippet, only title/url.
type Browserbase struct {
	Key    string
	Client *http.Client
	Count  int
}

func (b *Browserbase) Name() string { return "browserbase" }

type browserbaseRequest struct {
	Query      string `json:"query"`
	NumResults int    `json:"numResults,omitempty"`
}

type browserbaseResponse struct {
	Results []struct {
		Title string `json:"title"`
		URL   string `json:"url"`
	} `json:"results"`
}

func (b *Browserbase) Search(ctx context.Context, query string, count int) ([]Result, error) {
	if b.Key == "" {
		return nil, ErrBlocked
	}
	body, _ := json.Marshal(browserbaseRequest{
		Query: query, NumResults: countAt(count, b.Count),
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.browserbase.com/v1/search", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTransient, err)
	}
	httpc.Fingerprint(req.Header)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-bb-api-key", b.Key)
	var parsed browserbaseResponse
	if derr := decode(b.Client, req, &parsed, 0); derr != nil {
		return nil, derr
	}
	out := make([]Result, 0, len(parsed.Results))
	for _, r := range parsed.Results {
		out = append(out, Result{Title: r.Title, URL: r.URL, Engine: b.Name()})
	}
	return out, nil
}
