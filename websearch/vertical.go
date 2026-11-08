package websearch

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/FacileStudio/sonar/internal/engines"
)

// ErrNoVertical is returned by Images and News when no engine that serves
// that vertical is configured. Brave is the only one today.
var ErrNoVertical = errors.New("websearch: no engine serves this vertical")

// Image is one image search hit. URL is the page the image sits on and
// Thumbnail the provider's resized copy, which callers should proxy rather
// than hot-link.
type Image struct {
	Title     string
	URL       string
	Thumbnail string
	Source    string
	Width     int
	Height    int
	Engine    string
}

// News is one news search hit.
type News struct {
	Title     string
	URL       string
	Snippet   string
	Source    string
	Age       string
	Published string
	Thumbnail string
	Engine    string
}

func newBrave(key string, timeout time.Duration) *engines.Brave {
	if key == "" {
		return nil
	}
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	return &engines.Brave{Key: key, Client: &http.Client{Timeout: timeout}}
}

// Images runs an image search on Brave and returns at most count hits.
// Results are never cached.
func (c *Client) Images(ctx context.Context, query string, count int) ([]Image, error) {
	if c.brave == nil {
		return nil, ErrNoVertical
	}
	hits, err := c.brave.Images(ctx, query, count)
	if err != nil {
		return nil, err
	}
	out := make([]Image, 0, len(hits))
	for _, h := range hits {
		out = append(out, Image{Title: h.Title, URL: h.URL, Thumbnail: h.Thumbnail, Source: h.Source,
			Width: h.Width, Height: h.Height, Engine: "brave"})
	}
	return out, nil
}

// News runs a news search on Brave and returns at most count hits. Results
// are never cached.
func (c *Client) News(ctx context.Context, query string, count int) ([]News, error) {
	if c.brave == nil {
		return nil, ErrNoVertical
	}
	hits, err := c.brave.News(ctx, query, count)
	if err != nil {
		return nil, err
	}
	out := make([]News, 0, len(hits))
	for _, h := range hits {
		out = append(out, News{Title: h.Title, URL: h.URL, Snippet: h.Snippet, Source: h.Source,
			Age: h.Age, Published: h.Published, Thumbnail: h.Thumbnail, Engine: "brave"})
	}
	return out, nil
}
