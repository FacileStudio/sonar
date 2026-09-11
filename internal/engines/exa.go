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
	res, err := e.Client.Do(req)
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
	var parsed exaResponse
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("%w: decode: %v", ErrTransient, err)
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
