package engines

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBraveVerticalsParseImagesAndNews(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Subscription-Token") != "k" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch r.URL.Path {
		case "/res/v1/images/search":
			_, _ = w.Write([]byte(`{"results":[{"title":"a cat","url":"https://p.example/cat","source":"p.example",` +
				`"thumbnail":{"src":"https://imgs.example/t.jpg"},"properties":{"width":640,"height":480}}]}`))
		case "/res/v1/news/search":
			_, _ = w.Write([]byte(`{"results":[{"title":"cats win","url":"https://n.example/a","description":"d",` +
				`"age":"2 hours ago","page_age":"2026-11-08T10:00:00","meta_url":{"hostname":"n.example"}}]}`))
		}
	}))
	defer srv.Close()
	b := &Brave{Key: "k", Client: srv.Client(), BaseURL: srv.URL}
	images, err := b.Images(context.Background(), "cats", 5)
	if err != nil || len(images) != 1 || images[0].Thumbnail != "https://imgs.example/t.jpg" || images[0].Width != 640 {
		t.Fatalf("images = %+v, %v", images, err)
	}
	news, err := b.News(context.Background(), "cats", 5)
	if err != nil || len(news) != 1 || news[0].Source != "n.example" || news[0].Age != "2 hours ago" {
		t.Fatalf("news = %+v, %v", news, err)
	}
	if _, err := (&Brave{Key: "bad", Client: srv.Client(), BaseURL: srv.URL}).News(context.Background(), "x", 1); err == nil {
		t.Fatal("a refused key must fail")
	}
}

func TestPlainStripsHighlightsAndEntities(t *testing.T) {
	if got := Plain(`use <strong>adapter-static</strong> &amp; it&#x27;s fine`); got != "use adapter-static & it's fine" {
		t.Fatalf("Plain = %q", got)
	}
}
