package engines

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/FacileStudio/sonar/internal/httpc"
)

// BrightData queries the Bright Data SERP API — a keyed, provider-metred
// service that runs Google/Bing on its own infrastructure, so ruche's IP
// reputation never applies. The response envelope wraps the parsed JSON inside
// a stringified "body" field, and each organic hit's clean host lives in
// "display_link" (the "link" field is a Google redirect wrapper).
type BrightData struct {
	Key    string
	Zone   string
	Client *http.Client
	Count  int
}

func (b *BrightData) Name() string { return "brightdata" }

type brightdataRequest struct {
	Zone       string `json:"zone"`
	URL        string `json:"url"`
	Format     string `json:"format"`
	DataFormat string `json:"data_format,omitempty"`
}

type brightdataEnvelope struct {
	StatusCode int             `json:"status_code"`
	Body       json.RawMessage `json:"body"`
}

type brightdataOrganic struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	DisplayLink string `json:"display_link"`
	Description string `json:"description"`
}

type brightdataInner struct {
	Organic []brightdataOrganic `json:"organic"`
}

func (b *BrightData) Search(ctx context.Context, query string, count int) ([]Result, error) {
	if b.Key == "" || b.Zone == "" {
		return nil, ErrBlocked
	}
	searchURL := "https://www.google.com/search?" + url.Values{
		"q": {query},
	}.Encode()
	body, _ := json.Marshal(brightdataRequest{
		Zone: b.Zone, URL: searchURL, Format: "json",
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.brightdata.com/request", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTransient, err)
	}
	httpc.Fingerprint(req.Header)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+b.Key)
	res, err := b.Client.Do(req)
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
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: read: %v", ErrTransient, err)
	}
	inner, envStatus, err := decodeBrightData(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: parse: %v", ErrTransient, err)
	}
	// Bright Data returns HTTP 200 with a non-200 envelope status (e.g. 502
	// plus x-brd-error: captcha) when its scraper is blocked. Surface that as
	// transient so the dispatcher retries / falls through instead of a 0-hit.
	if envStatus != 0 && envStatus != http.StatusOK {
		return nil, fmt.Errorf("%w: upstream status %d", ErrTransient, envStatus)
	}
	out := make([]Result, 0, len(inner.Organic))
	for _, r := range inner.Organic {
		u := r.DisplayLink
		if u == "" {
			u = r.Link
		}
		out = append(out, Result{Title: r.Title, URL: u, Snippet: r.Description, Engine: b.Name()})
	}
	return out, nil
}

// decodeBrightData unwraps the SERP response and returns the envelope's
// upstream status (0 when the response was a flat object). Results live either
// at the object root or inside the envelope's "body" field, which is a JSON
// object in the parsed form and a stringified blob in the SDK form.
func decodeBrightData(raw []byte) (brightdataInner, int, error) {
	var inner brightdataInner
	if err := json.Unmarshal(raw, &inner); err != nil {
		return inner, 0, err
	}
	if len(inner.Organic) > 0 {
		return inner, 0, nil
	}
	var env brightdataEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return inner, 0, err
	}
	if len(env.Body) == 0 {
		return inner, env.StatusCode, nil
	}
	// A quoted-empty body (captcha/502 short-circuits) is no results, not an error.
	if string(env.Body) == `""` {
		return inner, env.StatusCode, nil
	}
	// "body" may be a JSON object or a stringified JSON blob; unwrap either.
	var asString string
	if err := json.Unmarshal(env.Body, &asString); err == nil {
		return inner, env.StatusCode, json.Unmarshal([]byte(asString), &inner)
	}
	return inner, env.StatusCode, json.Unmarshal(env.Body, &inner)
}
