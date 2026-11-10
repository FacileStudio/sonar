package mcp

import (
	"strings"
	"testing"

	"github.com/FacileStudio/sonar/websearch"
	"github.com/FacileStudio/tronc/mcp/mcptest"
)

func TestConformance(t *testing.T) {
	f := newFixture(t, -1, nil)
	mcptest.Run(t, mcptest.Config{URL: f.resource, Token: f.token})
}

func TestSearchSpendsQuotaThenNamesTheUpgradePage(t *testing.T) {
	f := newFixture(t, 1, nil)
	var out searchOutput
	mcptest.Decode(t, f.call(t, toolSearch, map[string]any{"query": "!g go generics"}), &out)
	if len(out.Results) != 1 || out.OwnKeys || f.quota.used != 1 || f.instance.queries[0] != "!g go generics" {
		t.Fatalf("want one literal search on sonar's keys that spends one search, got %+v %d %v", out, f.quota.used,
			f.instance.queries)
	}
	wantError(t, f.call(t, toolSearch, map[string]any{"query": "again"}), "Upgrade at https://sso.example/billing")
	wantError(t, f.call(t, toolFetch, map[string]any{"url": "https://example.com/"}), "Upgrade at")
}

func TestEnginesNeedOwnKeys(t *testing.T) {
	f := newFixture(t, -1, nil)
	wantError(t, f.call(t, toolSearch, map[string]any{"query": "q", "engines": []string{"exa"}}), "own engine keys")
	if len(f.instance.queries) != 0 {
		t.Fatal("a refused engine choice must not search")
	}
	f = newFixture(t, 0, keys{"exa": "k"})
	var out searchOutput
	mcptest.Decode(t, f.call(t, toolSearch, map[string]any{"query": "q", "engines": []string{"exa"}}), &out)
	if !out.OwnKeys || f.quota.used != 0 || len(f.own.queries) != 1 {
		t.Fatalf("own keys must search on them and spend nothing, got %+v %d", out, f.quota.used)
	}
}

func TestSearchArguments(t *testing.T) {
	f := newFixture(t, -1, nil)
	wantError(t, f.call(t, toolSearch, map[string]any{"query": "q", "kind": "video"}), "kind must be")
	wantError(t, f.call(t, toolSearch, map[string]any{"query": "q", "count": 50}), "count must be")
	wantError(t, f.call(t, toolSearch, map[string]any{"query": " "}), "Q is required")
	var out searchOutput
	mcptest.Decode(t, f.call(t, toolSearch, map[string]any{"query": "cat", "kind": "images"}), &out)
	if len(out.Images) != 1 || out.Images[0].ImageURL != "https://imgs.example/t.jpg" {
		t.Fatalf("image results must carry the image url, not the session-only proxy, got %+v", out.Images)
	}
	if f.quota.used != 1 {
		t.Fatalf("refused arguments must not spend quota, used %d", f.quota.used)
	}
}

func TestFetchReadsLongPagesInParts(t *testing.T) {
	f := newFixture(t, -1, nil)
	f.instance.markdown = strings.Repeat("é", 20_000)
	var first, second fetchOutput
	mcptest.Decode(t, f.call(t, toolFetch, map[string]any{"url": "https://example.com/"}), &first)
	if first.NextOffset == 0 || len(first.Markdown) > defaultLength || first.TotalLength != 40_000 {
		t.Fatalf("want a first part and a next_offset, got %d bytes, next %d", len(first.Markdown), first.NextOffset)
	}
	mcptest.Decode(t, f.call(t, toolFetch, map[string]any{"url": "https://example.com/", "offset": first.NextOffset + 1}),
		&second)
	if first.Markdown+second.Markdown != f.instance.markdown || second.NextOffset != 0 {
		t.Fatalf("the parts must join into the page on rune boundaries, got %d + %d bytes", len(first.Markdown),
			len(second.Markdown))
	}
}

func TestFetchRefusesInternalAddresses(t *testing.T) {
	f := newFixture(t, -1, nil)
	f.instance.fail = &websearch.BlockedError{}
	wantError(t, f.call(t, toolFetch, map[string]any{"url": "http://10.0.0.1/"}), "not on the public internet")
	wantError(t, f.call(t, toolFetch, map[string]any{"url": "file:///etc/passwd"}), "http or https")
	if f.quota.used != 0 {
		t.Fatalf("a refused fetch must not spend quota, used %d", f.quota.used)
	}
}
