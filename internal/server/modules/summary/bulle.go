package summary

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const actingForHeader = "X-Bulle-Acting-For"

// Bulle writes summaries through Bulle's service API. BaseURL is fixed by the
// operator and Key is sonar's service key in Bulle's BULLE_AI_CALLERS.
//
// The route it calls, POST /api/ai/summarize, does not exist in Bulle yet; the
// contract sonar expects is documented in docs/ai-summary.md.
type Bulle struct {
	BaseURL string
	Key     string
	Client  *http.Client
}

// NewBulle builds the Bulle writer, or nil when either setting is missing.
func NewBulle(baseURL, key string) *Bulle {
	if baseURL == "" || key == "" {
		return nil
	}
	return &Bulle{BaseURL: strings.TrimSuffix(baseURL, "/"), Key: key, Client: &http.Client{Timeout: 30 * time.Second}}
}

// Summarize asks Bulle to summarize req for the person behind actingFor.
func (b *Bulle) Summarize(ctx context.Context, actingFor string, req Request) (Response, error) {
	payload, err := json.Marshal(req)
	if err != nil {
		return Response{}, err
	}
	call, err := http.NewRequestWithContext(ctx, http.MethodPost, b.BaseURL+"/api/ai/summarize", bytes.NewReader(payload))
	if err != nil {
		return Response{}, err
	}
	call.Header.Set("Authorization", "Bearer "+b.Key)
	call.Header.Set("Content-Type", "application/json")
	call.Header.Set(actingForHeader, actingFor)
	res, err := b.Client.Do(call)
	if err != nil {
		return Response{}, err
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		return Response{}, fmt.Errorf("bulle answered %d", res.StatusCode)
	}
	var out Response
	if err := json.NewDecoder(io.LimitReader(res.Body, 64<<10)).Decode(&out); err != nil {
		return Response{}, err
	}
	return out, nil
}
