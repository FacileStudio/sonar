package engines

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/FacileStudio/sonar/internal/httpc"
)

// Linkup queries the Linkup Search API — a keyed, provider-metred web-search
// service tuned for AI retrieval, so ruche's IP reputation never applies.
// Returns ranked URLs with an extracted content snippet.
type Linkup struct {
	Key    string
	Client *http.Client
	Count  int
}

func (l *Linkup) Name() string { return "linkup" }

type linkupRequest struct {
	Q          string `json:"q"`
	Depth      string `json:"depth"`
	OutputType string `json:"outputType"`
	MaxResults int    `json:"maxResults"`
}

type linkupResponse struct {
	Results []struct {
		Name    string `json:"name"`
		URL     string `json:"url"`
		Content string `json:"content"`
	} `json:"results"`
}

func (l *Linkup) Search(ctx context.Context, query string, count int) ([]Result, error) {
	if l.Key == "" {
		return nil, ErrBlocked
	}
	body, _ := json.Marshal(linkupRequest{
		Q:          query,
		Depth:      "fast",
		OutputType: "searchResults",
		MaxResults: countAt(count, l.Count),
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.linkup.so/v1/search", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTransient, err)
	}
	httpc.Fingerprint(req.Header)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+l.Key)
	res, err := l.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTransient, err)
	}
	defer res.Body.Close()
	switch {
	case res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden || res.StatusCode == http.StatusPaymentRequired:
		return nil, ErrBlocked
	case res.StatusCode == http.StatusTooManyRequests || res.StatusCode >= 500:
		return nil, ErrTransient
	case res.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("%w: status %d", ErrTransient, res.StatusCode)
	}
	var parsed linkupResponse
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("%w: decode: %v", ErrTransient, err)
	}
	out := make([]Result, 0, len(parsed.Results))
	for _, r := range parsed.Results {
		out = append(out, Result{Title: r.Name, URL: r.URL, Snippet: r.Content, Engine: l.Name()})
	}
	return out, nil
}