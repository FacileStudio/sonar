// Package merge deduplicates engine results and ranks them so a single final
// list always comes out, whatever combination of engines answered.
package merge

import (
	"net/url"
	"sort"
	"strings"

	"github.com/FacileStudio/sonar/internal/engines"
)

// engineTrust weights an engine's results: keyed APIs and the local SearXNG
// aggregator rank above scrapers, whose relevance is untrustworthy from
// low-reputation IPs.
var engineTrust = map[string]int{
	"brave":       5,
	"brightdata":  5,
	"browserbase": 5,
	"exa":         5,
	"firecrawl":   5,
	"linkup":      5,
	"serpapi":     5,
	"tavily":      5,
	"furet":       4,
	"bing":        1,
	"ddg":         1,
}

// Rank dedupes by normalized URL and sorts by engine trust, then original
// position (which already encodes per-engine rank). Capped at count.
func Rank(in []engines.Result, count int) []engines.Result {
	seen := map[string]bool{}
	out := make([]engines.Result, 0, len(in))
	for _, r := range in {
		k := urlKey(r.URL)
		if k == "" || seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, r)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return trust(out[i].Engine) > trust(out[j].Engine)
	})
	if len(out) > count {
		out = out[:count]
	}
	return out
}

func trust(engine string) int {
	if w, ok := engineTrust[engine]; ok {
		return w
	}
	return 0
}

// urlKey normalizes a URL for dedup: scheme and www dropped, host and path
// lowercased, query and fragment stripped, trailing slash removed.
func urlKey(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return ""
	}
	if u.Host == "" {
		return ""
	}
	host := strings.TrimPrefix(strings.ToLower(u.Host), "www.")
	path := strings.ToLower(strings.TrimSuffix(u.Path, "/"))
	if path == "" {
		path = "/"
	}
	return host + path
}
