package engines

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/FacileStudio/sonar/internal/httpc"
)

// Exa queries the Exa API — a keyed semantic/neural search engine tuned for
// finding what content actually says, not just matching keywords. Unlike the
// free search APIs it is provider-metred, so ruche's IP reputation never
// applies.
type Exa struct {
	Key    string
	Client *http.Client
	Count  int
}

func (e *Exa) Name() string { return "exa" }

type exaRequest struct {
	Query      string `json:"query"`
	NumResults int    `json:"numResults"`
	Contents   struct {
		Text bool `json:"text"`
	} `json:"contents"`
}

type exaResponse struct {
	Results []struct {
		Title string `json:"title"`
		URL   string `json:"url"`
		Text  string `json:"text"`
	} `json:"results"`
}

func (e *Exa) Search(ctx context.Context, query string, count int) ([]Result, error) {
	if e.Key == "" {
		return nil, ErrBlocked
	}
	body, _ := json.Marshal(exaRequest{
		Query:      query,
		NumResults: countAt(count, e.Count),
		Contents: struct {
			Text bool `json:"text"`
		}{Text: true},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.exa.ai/search", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTransient, err)
	}
	httpc.Fingerprint(req.Header)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", e.Key)
	var parsed exaResponse
	if derr := decode(e.Client, req, &parsed, 0); derr != nil {
		return nil, derr
	}
	out := make([]Result, 0, len(parsed.Results))
	for _, r := range parsed.Results {
		out = append(out, Result{Title: r.Title, URL: r.URL, Snippet: truncate(r.Text), Engine: e.Name()})
	}
	return out, nil
}

func truncate(s string) string {
	const max = 300
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
