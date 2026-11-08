// Package search serves sonar's JSON search API: web, images and news
// searches, page fetches and suggestions. Queries are never logged or stored;
// the only trace a search leaves is one tick on the user's daily counter.
package search

import (
	"context"
	"time"

	"github.com/FacileStudio/sonar/websearch"
)

// Kind is a search tab.
type Kind string

// The tabs sonar serves.
const (
	KindWeb    Kind = "web"
	KindImages Kind = "images"
	KindNews   Kind = "news"
)

// Engines is what a search runs on: sonar's own keys, or a person's.
// *websearch.Client satisfies it.
type Engines interface {
	Search(ctx context.Context, query string, count int, include []string) ([]websearch.Result, error)
	Images(ctx context.Context, query string, count int) ([]websearch.Image, error)
	News(ctx context.Context, query string, count int) ([]websearch.News, error)
	Fetch(ctx context.Context, rawURL string, opts websearch.FetchOptions) (*websearch.Page, error)
}

// Plans answers a user's daily quota, -1 meaning unlimited.
type Plans interface {
	Daily(ctx context.Context, userID int64) (int64, error)
	CheckAI(ctx context.Context, userID int64) error
	Refusal(message string, cause error) error
}

// Quota is the daily counter.
type Quota interface {
	Take(ctx context.Context, userID, limit int64) (bool, error)
	Refund(ctx context.Context, userID int64) error
	Used(ctx context.Context, userID int64) (int64, error)
	ResetsAt() time.Time
}

// Keys opens the engine keys a user brought.
type Keys interface {
	Keys(ctx context.Context, userID int64) (map[string]string, error)
}

// WebResult is one ranked web hit.
type WebResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
	Engine  string `json:"engine"`
}

// ImageResult is one image hit. Thumbnail is a signed same-origin /img URL.
type ImageResult struct {
	Title     string `json:"title"`
	URL       string `json:"url"`
	Thumbnail string `json:"thumbnail"`
	Source    string `json:"source"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
}

// NewsResult is one news hit. Thumbnail is a signed same-origin /img URL.
type NewsResult struct {
	Title     string `json:"title"`
	URL       string `json:"url"`
	Snippet   string `json:"snippet"`
	Source    string `json:"source"`
	Age       string `json:"age,omitempty"`
	Published string `json:"published,omitempty"`
	Thumbnail string `json:"thumbnail,omitempty"`
}

// Response is the answer to GET /api/search. Redirect is set instead of
// results when the query carried an external !bang.
type Response struct {
	Query    string        `json:"query"`
	Kind     Kind          `json:"kind"`
	Redirect string        `json:"redirect,omitempty"`
	OwnKeys  bool          `json:"own_keys"`
	Web      []WebResult   `json:"web,omitempty"`
	Images   []ImageResult `json:"images,omitempty"`
	News     []NewsResult  `json:"news,omitempty"`
}

// Usage is the answer to GET /api/usage. Limit -1 is unlimited.
type Usage struct {
	Used     int64     `json:"used"`
	Limit    int64     `json:"limit"`
	ResetsAt time.Time `json:"resets_at"`
	OwnKeys  []string  `json:"own_keys"`
	AI       bool      `json:"ai"`
}

// FetchRequest is the body of POST /api/fetch.
type FetchRequest struct {
	URL string `json:"url"`
}

// FetchResponse is one fetched page as markdown.
type FetchResponse struct {
	URL       string `json:"url"`
	FinalURL  string `json:"final_url"`
	Title     string `json:"title"`
	Markdown  string `json:"markdown"`
	Truncated bool   `json:"truncated"`
}
