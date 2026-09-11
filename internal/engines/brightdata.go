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
	req := brightDataRequest(ctx, b, body)
	res, err := b.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTransient, err)
	}
	defer res.Body.Close()
	if gerr := guardStatus(res.StatusCode, 0); gerr != nil {
		return nil, gerr
	}
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: read: %v", ErrTransient, err)
	}
	inner, envStatus, err := decodeBrightData(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: parse: %v", ErrTransient, err)
	}
	return brightDataHits(inner, envStatus, b.Name())
}

// brightDataRequest builds the authenticated SERP POST request.
func brightDataRequest(ctx context.Context, b *BrightData, body []byte) *http.Request {
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.brightdata.com/request", bytes.NewReader(body))
	httpc.Fingerprint(req.Header)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+b.Key)
	return req
}

// brightDataHits flattens the envelope's organic results, or an upstream error
// when the envelope carries a non-OK status (e.g. 502 plus a captcha header).
func brightDataHits(inner brightdataInner, envStatus int, engine string) ([]Result, error) {
	if envStatus != 0 && envStatus != http.StatusOK {
		return nil, fmt.Errorf("%w: upstream status %d", ErrTransient, envStatus)
	}
	out := make([]Result, 0, len(inner.Organic))
	for _, r := range inner.Organic {
		u := r.DisplayLink
		if u == "" {
			u = r.Link
		}
		out = append(out, Result{Title: r.Title, URL: u, Snippet: r.Description, Engine: engine})
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
	if string(env.Body) == `""` {
		return inner, env.StatusCode, nil
	}
	var asString string
	if err := json.Unmarshal(env.Body, &asString); err == nil {
		return inner, env.StatusCode, json.Unmarshal([]byte(asString), &inner)
	}
	return inner, env.StatusCode, json.Unmarshal(env.Body, &inner)
}
