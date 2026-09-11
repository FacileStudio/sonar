package search

import (
	"net/http"

	"github.com/FacileStudio/sonar/internal/config"
	"github.com/FacileStudio/sonar/internal/engines"
)

type built struct {
	prio int
	eng  engines.Engine
}

// buildEngineUnits turns the config map into ordered engine instances. Only
// enabled engines are built; disabled or keyed-but-missing-key entries are
// skipped so the dispatcher never calls a worthless backend.
func buildEngineUnits(cfg *config.Config, client *http.Client) []built {
	var out []built
	for name, def := range cfg.Engines {
		if def == nil || !def.Enabled {
			continue
		}
		if e := makeEngine(name, def, client); e != nil {
			out = append(out, built{prio: def.Priority, eng: e})
		}
	}
	return out
}

// needsKey lists the engines that require a provider API key to be built. The
// scrape engines (furet, bing, ddg) never need one.
var needsKey = map[string]bool{
	"brave": true, "tavily": true, "exa": true, "firecrawl": true, "serpapi": true,
	"browserbase": true, "brightdata": true, "linkup": true,
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

// makeKeyed builds one of the eight keyed-API engines, or nil for a scrape name.
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

// makeScrape builds an engine that needs no API key: furet, bing or ddg.
func makeScrape(name string, def *config.EngineDef, client *http.Client) engines.Engine {
	switch name {
	case "furet":
		base := def.BaseURL
		if base == "" {
			base = "https://furet.facile.studio"
		}
		return &engines.Furet{BaseURL: base, Client: client, Count: def.Count}
	case "bing":
		return &engines.Bing{Client: client, Count: def.Count}
	case "ddg":
		return &engines.DDG{Client: client, Count: def.Count}
	}
	return nil
}
