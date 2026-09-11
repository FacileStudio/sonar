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
	// "dup" and "low1" share the same normalized URL, so 2 unique survive.
	if len(out) != 2 {
		t.Fatalf("dedup failed: got %d results, want 2", len(out))
	}
	// the trusted engine's result must outrank the scraper's.
	if out[0].Title != "high" {
		t.Fatalf("expected trusted engine first, got %q", out[0].Title)
	}
}

func TestRankCapsAtCount(t *testing.T) {
	in := []engines.Result{
		{Title: "a", URL: "https://x.com/1", Engine: "furet"},
		{Title: "b", URL: "https://x.com/2", Engine: "furet"},
		{Title: "c", URL: "https://x.com/3", Engine: "furet"},
	}
	if got := len(Rank(in, 2)); got != 2 {
		t.Fatalf("count cap failed: got %d, want 2", got)
	}
}
