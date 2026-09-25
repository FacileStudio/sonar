package engines

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestExtractPagesEmpty(t *testing.T) {
	_, err := ExtractPages(context.Background(), nil, ExtractOptions{})
	if err == nil {
		t.Fatal("expected error on empty URLs")
	}
}

func TestExtractHTTPMock(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `<!DOCTYPE html><html><head><title>Test Title</title></head><body><h1>Heading</h1><p>` +
			`This is a sufficiently long body text that ensures the markdown conversion ` +
			`produces more than 150 characters of useful content for testing purposes. ` +
			`It contains enough words to satisfy the length threshold.</p></body></html>`
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, html)
	}))
	defer srv.Close()

	res, err := extractHTTP(context.Background(), srv.URL, srv.Client())
	if err != nil {
		t.Fatalf("unexpected extract error: %v", err)
	}
	if res.Title != "Test Title" {
		t.Fatalf("expected Test Title, got %q", res.Title)
	}
	if len(res.Snippet) < 100 {
		t.Fatalf("snippet too short: %q", res.Snippet)
	}
}

func TestExtractPagesHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `<html><head><title>Page</title></head><body><p>` +
			`Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor ` +
			`incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud.</p></body></html>`
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, html)
	}))
	defer srv.Close()

	urls := []string{srv.URL + "/1", srv.URL + "/2"}
	opts := ExtractOptions{Timeout: 5 * time.Second}
	results, err := ExtractPages(context.Background(), urls, opts)
	if err != nil {
		t.Fatalf("ExtractPages error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}
