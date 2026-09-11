package merge

import (
	"testing"

	"github.com/FacileStudio/sonar/internal/engines"
)

func TestRankDedupAndTrust(t *testing.T) {
	in := []engines.Result{
		{Title: "low1", URL: "https://www.example.com/a", Engine: "bing"},
		{Title: "dup", URL: "https://example.com/A?ref=1", Engine: "ddg"},
		{Title: "high", URL: "https://example.com/b", Engine: "brave"},
	}
	out := Rank(in, 10)
	if len(out) != 2 {
		t.Fatalf("dedup failed: got %d results, want 2", len(out))
	}
	if out[0].Title != "high" {
		t.Fatalf("expected trusted engine first, got %q", out[0].Title)
	}
}

func TestRankCapsAtCount(t *testing.T) {
	in := []engines.Result{
		{Title: "a", URL: "https://x.com/1", Engine: "searxng"},
		{Title: "b", URL: "https://x.com/2", Engine: "searxng"},
		{Title: "c", URL: "https://x.com/3", Engine: "searxng"},
	}
	if got := len(Rank(in, 2)); got != 2 {
		t.Fatalf("count cap failed: got %d, want 2", got)
	}
}

func TestTrustSearxngSitsAboveScrapersBelowKeyed(t *testing.T) {
	if trust("searxng") <= trust("bing") {
		t.Fatal("searxng should outrank a scraper")
	}
	if trust("searxng2") <= trust("bing") {
		t.Fatal("indexed searxng instance should outrank a scraper")
	}
	if trust("searxng") >= trust("brave") {
		t.Fatal("searxng should rank below a keyed API")
	}
}
