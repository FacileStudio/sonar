package engines

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// Image is one image search hit. URL is the page the image sits on, Thumbnail
// the provider's resized copy and Source the host that published it.
type Image struct {
	Title     string
	URL       string
	Thumbnail string
	Source    string
	Width     int
	Height    int
}

// News is one news search hit. Age is the provider's human-readable age, such
// as "2 hours ago", and Published the article time when the provider gives one.
type News struct {
	Title     string
	URL       string
	Snippet   string
	Source    string
	Age       string
	Published string
	Thumbnail string
}

type braveImages struct {
	Results []struct {
		Title     string `json:"title"`
		URL       string `json:"url"`
		Source    string `json:"source"`
		Thumbnail struct {
			Src string `json:"src"`
		} `json:"thumbnail"`
		Properties struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"properties"`
	} `json:"results"`
}

type braveNews struct {
	Results []struct {
		Title       string `json:"title"`
		URL         string `json:"url"`
		Description string `json:"description"`
		Age         string `json:"age"`
		PageAge     string `json:"page_age"`
		MetaURL     struct {
			Hostname string `json:"hostname"`
		} `json:"meta_url"`
		Thumbnail struct {
			Src string `json:"src"`
		} `json:"thumbnail"`
	} `json:"results"`
}

// Images runs a Brave image search. Results are never cached.
func (b *Brave) Images(ctx context.Context, query string, count int) ([]Image, error) {
	var parsed braveImages
	if err := braveVertical(ctx, b, vertical{"images", query, min(countAt(count, 0), 100)}, &parsed); err != nil {
		return nil, err
	}
	out := make([]Image, 0, len(parsed.Results))
	for _, r := range parsed.Results {
		out = append(out, Image{
			Title: r.Title, URL: r.URL, Thumbnail: r.Thumbnail.Src, Source: r.Source,
			Width: r.Properties.Width, Height: r.Properties.Height,
		})
	}
	return out, nil
}

// News runs a Brave news search. Results are never cached.
func (b *Brave) News(ctx context.Context, query string, count int) ([]News, error) {
	var parsed braveNews
	if err := braveVertical(ctx, b, vertical{"news", query, min(countAt(count, 0), 50)}, &parsed); err != nil {
		return nil, err
	}
	out := make([]News, 0, len(parsed.Results))
	for _, r := range parsed.Results {
		out = append(out, News{
			Title: r.Title, URL: r.URL, Snippet: r.Description, Source: r.MetaURL.Hostname,
			Age: r.Age, Published: r.PageAge, Thumbnail: r.Thumbnail.Src,
		})
	}
	return out, nil
}

type vertical struct {
	kind  string
	query string
	count int
}

func braveVertical[T any](ctx context.Context, b *Brave, v vertical, out *T) error {
	if b.Key == "" {
		return ErrBlocked
	}
	u := b.base() + "/res/v1/" + v.kind + "/search?" + url.Values{
		"q":     {v.query},
		"count": {fmt.Sprint(v.count)},
	}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Subscription-Token", b.Key)
	req.Header.Set("Accept", "application/json")
	return decode(b.Client, req, out, 0)
}
