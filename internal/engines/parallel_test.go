package engines

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParallelName(t *testing.T) {
	p := &Parallel{}
	if p.Name() != "parallel" {
		t.Fatalf("expected parallel, got %s", p.Name())
	}
}

func TestParallelMissingKey(t *testing.T) {
	p := &Parallel{}
	if _, err := p.Search(context.Background(), "query", 5); err == nil {
		t.Fatal("expected error for missing key")
	}
	if _, err := p.Extract(context.Background(), []string{"https://example.com"}); err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestParallelSearchMock(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"results":[{"url":"https://example.com","title":"Example","excerpts":["test"]}]}`)
	}))
	defer srv.Close()

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodPost, srv.URL, nil)
	var parsed parallelSearchResponse
	if err := decode(srv.Client(), req, &parsed, 0); err != nil {
		t.Fatalf("unexpected decode error: %v", err)
	}
	if len(parsed.Results) != 1 || parsed.Results[0].Title != "Example" {
		t.Fatalf("unexpected results: %+v", parsed)
	}
}
