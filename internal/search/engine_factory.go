package search

import (
	"fmt"
	"net/http"

	"github.com/FacileStudio/sonar/internal/config"
	"github.com/FacileStudio/sonar/internal/engines"
)

type built struct {
	name   string
	prio   int
	eng    engines.Engine
	scrape bool
}

// buildEngineUnits turns the config map and the user's SearXNG list into
// ordered engine instances. Only enabled engines are built; disabled or
// keyed-but-missing-key entries are skipped so the dispatcher never calls a
// worthless backend.
func buildEngineUnits(cfg *config.Config, client *http.Client) []built {
	var out []built
	for name, def := range cfg.Engines {
		if def == nil || !def.Enabled {
			continue
		}
		if e := makeEngine(name, def, client); e != nil {
			out = append(out, built{name: name, prio: def.Priority, eng: e, scrape: isScrape(name)})
		}
	}
	for i, base := range cfg.Searxng {
		out = append(out, built{
			name: SearxngName(cfg.Searxng, i),
			prio: cfg.SearxngPriority,
			eng:  &engines.Searxng{Label: SearxngName(cfg.Searxng, i), BaseURL: base, Client: client},
		})
	}
	return out
}

// SearxngName labels a SearXNG instance for the cache and merge layers: the
// bare name for a single instance, an index suffix when there are several so
// their caches and breaker identities stay distinct.
func SearxngName(urls []string, i int) string {
	if len(urls) > 1 {
		return fmt.Sprintf("searxng%d", i+1)
	}
	return "searxng"
}

// needsKey lists the engines that require a provider API key to be built. The
// scrape engines (bing, ddg) and SearXNG instances never need one.
var needsKey = map[string]bool{
	"brave": true, "tavily": true, "exa": true, "firecrawl": true, "serpapi": true,
	"browserbase": true, "brightdata": true, "linkup": true,
	"rod": false,
}

// NeedsKey reports whether an engine needs a provider API key to build. Used by
// the config command to tell a keyed-but-unkeyed engine apart from a scrape one.
func NeedsKey(name string) bool {
	return needsKey[name]
}

// makeEngine builds one engine from its config, or nil when it is a keyed
// engine with no key — an entry the dispatcher would waste a call on.
func makeEngine(name string, def *config.EngineDef, client *http.Client) engines.Engine {
	if needsKey[name] && def.APIKey == "" {
		return nil
	}
	if e := makeKeyed(name, def, client); e != nil {
		return e
	}
	return makeScrape(name, def, client)
}

// isScrape reports whether an engine is scraped from a shared search endpoint
// and must therefore never run concurrently with another scraper.
func isScrape(name string) bool {
	return name == "bing" || name == "ddg"
}

// makeKeyed builds one of the keyed-API or browser engines, or nil for a scrape name.
func makeKeyed(name string, def *config.EngineDef, client *http.Client) engines.Engine {
	switch name {
	case "brave":
		return &engines.Brave{Key: def.APIKey, Client: client, Count: def.Count}
	case "tavily":
		return &engines.Tavily{Key: def.APIKey, Client: client, Count: def.Count}
	case "exa":
		return &engines.Exa{Key: def.APIKey, Client: client, Count: def.Count}
	case "firecrawl":
		return &engines.Firecrawl{Key: def.APIKey, Client: client, Count: def.Count}
	case "serpapi":
		return &engines.SerpAPI{Key: def.APIKey, Client: client, Count: def.Count}
	case "browserbase":
		return &engines.Browserbase{Key: def.APIKey, Client: client, Count: def.Count}
	case "brightdata":
		return &engines.BrightData{Key: def.APIKey, Zone: def.Zone, Client: client, Count: def.Count}
	case "linkup":
		return &engines.Linkup{Key: def.APIKey, Client: client, Count: def.Count}
	}
	return nil
}

// makeScrape builds an engine that needs no API key: bing, ddg, or rod.
func makeScrape(name string, def *config.EngineDef, client *http.Client) engines.Engine {
	switch name {
	case "bing":
		return &engines.Bing{Client: client, Count: def.Count}
	case "ddg":
		return &engines.DDG{Client: client, Count: def.Count}
	case "rod":
		return &engines.Rod{Count: def.Count}
	}
	return nil
}
