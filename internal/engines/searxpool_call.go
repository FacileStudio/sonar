package engines

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
)

const (
	poolUserAgent = "sonar-search (+https://sonar.facile.studio)"
	poolMaxBody   = 2 << 20
)

// call asks one instance for JSON results. It declares itself plainly, never
// retries and never tries to pass a bot wall: a refusal is reported and the
// caller parks the instance. Errors name no URL and no query text.
func (p *SearxPool) call(ctx context.Context, base, query string) ([]Result, error) {
	form := url.Values{"q": {query}, "format": {"json"}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/search", strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("%w: bad instance address", ErrTransient)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", poolUserAgent)
	res, err := p.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: request failed", ErrTransient)
	}
	defer res.Body.Close()
	if gerr := poolStatus(res.StatusCode); gerr != nil {
		return nil, gerr
	}
	return decodePool(res)
}

func poolStatus(code int) error {
	switch {
	case code == http.StatusOK:
		return nil
	case code == http.StatusForbidden || code == http.StatusTooManyRequests:
		return ErrBlocked
	}
	return fmt.Errorf("%w: status %d", ErrTransient, code)
}

func decodePool(res *http.Response) ([]Result, error) {
	kind, _, _ := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if kind != "application/json" {
		return nil, ErrBlocked
	}
	var parsed searxResponse
	if err := json.NewDecoder(io.LimitReader(res.Body, poolMaxBody)).Decode(&parsed); err != nil {
		return nil, ErrBlocked
	}
	out := make([]Result, 0, len(parsed.Results))
	for _, r := range parsed.Results {
		if r.URL != "" && r.Title != "" {
			out = append(out, Result{Title: r.Title, URL: r.URL, Snippet: r.Content, Engine: PoolName})
		}
	}
	return out, nil
}
