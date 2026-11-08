package search

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"
)

// Suggester completes a partial query from an upstream suggestion endpoint
// that answers the OpenSearch suggestions format, DuckDuckGo's by default. The
// partial query goes upstream with no cookie, account or address of the
// person typing; nothing is logged or kept.
type Suggester struct {
	Endpoint string
	Client   *http.Client
}

// NewSuggester builds a Suggester for endpoint, or nil when endpoint is empty.
func NewSuggester(endpoint string) *Suggester {
	if endpoint == "" {
		return nil
	}
	return &Suggester{Endpoint: endpoint, Client: &http.Client{Timeout: 3 * time.Second}}
}

// Complete returns at most eight completions for partial. Any upstream failure
// is an empty list: suggestions are a convenience, never an error.
func (s *Suggester) Complete(ctx context.Context, partial string) []string {
	if s == nil || partial == "" {
		return []string{}
	}
	target := s.Endpoint + "?" + url.Values{"q": {partial}, "type": {"list"}}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return []string{}
	}
	req.Header.Set("Accept", "application/json")
	res, err := s.Client.Do(req)
	if err != nil {
		return []string{}
	}
	defer func() { _ = res.Body.Close() }()
	var body []json.RawMessage
	if res.StatusCode != http.StatusOK || json.NewDecoder(io.LimitReader(res.Body, 64<<10)).Decode(&body) != nil ||
		len(body) < 2 {
		return []string{}
	}
	var list []string
	if json.Unmarshal(body[1], &list) != nil {
		return []string{}
	}
	return list[:min(len(list), 8)]
}

func (h *handler) suggestions(w http.ResponseWriter, r *http.Request) {
	partial := strings.TrimSpace(r.URL.Query().Get("q"))
	if utf8.RuneCountInString(partial) > maxQuery {
		partial = ""
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/x-suggestions+json; charset=utf-8")
	_ = json.NewEncoder(w).Encode([]any{partial, h.suggest.Complete(r.Context(), partial)})
}
