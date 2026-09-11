package engines

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/FacileStudio/sonar/internal/httpc"
)

// BrightData queries the Bright Data SERP API — a keyed, provider-metred
// service that runs Google/Bing on its own infrastructure, so ruche's IP
// reputation never applies. It returns structured organic results.
type BrightData struct {
	Key    string
	Zone   string
	Client *http.Client
	Count  int
}

func (b *BrightData) Name() string { return "brightdata" }

type brightdataRequest struct {
	Zone   string `json:"zone"`
	URL    string `json:"url"`
	Format string `json:"format"`
}

type brightdataResponse struct {
	Organic []struct {
		Title       string `json:"title"`
		Link        string `json:"link"`
		Description string `json:"description"`
	} `json:"organic"`
}

func (b *BrightData) Search(ctx context.Context, query string, count int) ([]Result, error) {
	if b.Key == "" || b.Zone == "" {
		return nil, ErrBlocked
	}
	n := countAt(count, b.Count)
	searchURL := "https://www.google.com/search?" + url.Values{
		"q": {query}, "num": {fmt.Sprint(n)},
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
	var parsed brightdataResponse
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("%w: decode: %v", ErrTransient, err)
	}
	out := make([]Result, 0, len(parsed.Organic))
	for _, r := range parsed.Organic {
		out = append(out, Result{Title: r.Title, URL: r.Link, Snippet: r.Description, Engine: b.Name()})
	}
	return out, nil
}