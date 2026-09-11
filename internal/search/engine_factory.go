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
		switch name {
		case "brave":
			if def.APIKey == "" {
				continue
			}
			out = append(out, built{prio: def.Priority, eng: &engines.Brave{Key: def.APIKey, Client: client, Count: def.Count}})
		case "tavily":
			if def.APIKey == "" {
				continue
			}
			out = append(out, built{prio: def.Priority, eng: &engines.Tavily{Key: def.APIKey, Client: client, Count: def.Count}})
		case "exa":
			if def.APIKey == "" {
				continue
			}
			out = append(out, built{prio: def.Priority, eng: &engines.Exa{Key: def.APIKey, Client: client, Count: def.Count}})
		case "firecrawl":
			if def.APIKey == "" {
				continue
			}
			out = append(out, built{prio: def.Priority, eng: &engines.Firecrawl{Key: def.APIKey, Client: client, Count: def.Count}})
		case "serpapi":
			if def.APIKey == "" {
				continue
			}
			out = append(out, built{prio: def.Priority, eng: &engines.SerpAPI{Key: def.APIKey, Client: client, Count: def.Count}})
		case "browserbase":
			if def.APIKey == "" {
				continue
			}
			out = append(out, built{prio: def.Priority, eng: &engines.Browserbase{Key: def.APIKey, Client: client, Count: def.Count}})
		case "brightdata":
			if def.APIKey == "" {
				continue
			}
			out = append(out, built{prio: def.Priority, eng: &engines.BrightData{Key: def.APIKey, Zone: def.Zone, Client: client, Count: def.Count}})
		case "linkup":
			if def.APIKey == "" {
				continue
			}
			out = append(out, built{prio: def.Priority, eng: &engines.Linkup{Key: def.APIKey, Client: client, Count: def.Count}})
		case "furet":
			base := def.BaseURL
			if base == "" {
				base = "https://furet.facile.studio"
			}
			out = append(out, built{prio: def.Priority, eng: &engines.Furet{BaseURL: base, Client: client, Count: def.Count}})
		case "bing":
			out = append(out, built{prio: def.Priority, eng: &engines.Bing{Client: client, Count: def.Count}})
		case "ddg":
			out = append(out, built{prio: def.Priority, eng: &engines.DDG{Client: client, Count: def.Count}})
		}
	}
	return out
}
