// Package server is `sonar serve`: the hosted sonar at sonar.facile.studio,
// with porte sign-in against Registre, the JSON search API, the image proxy,
// per-user daily quotas from plan entitlements and the SvelteKit client.
//
// /mcp serves the same search service to MCP clients, behind Registre
// access tokens bound to MCP_RESOURCE.
package server

import (
	"fmt"
	"strings"

	"github.com/FacileStudio/porte"
	portepg "github.com/FacileStudio/porte/pg"
	"github.com/FacileStudio/sonar/internal/server/modules/search"
	troncenv "github.com/FacileStudio/tronc/env"
)

// instanceEngines are the keyed engines sonar runs on its own keys, each read
// from SONAR_<NAME>_KEY.
var instanceEngines = []string{"brave", "exa", "firecrawl", "linkup", "parallel", "serpapi", "tavily"}

// Engines are sonar's own engines and the policy it runs them under.
type Engines struct {
	Keys    map[string]string
	Searxng []string
	Web     search.Chain
	Sources search.Chain
}

// Config is the process configuration.
type Config struct {
	troncenv.Core
	PublicURL    string
	OIDC         porte.Config
	ClientSecret string
	Sealer       *portepg.TokenSealer
	Engines      Engines
	SuggestURL   string
	ImageKey     []byte
	Summary      bool
	BulleURL     string
	BulleKey     string
	MCPResource  string
}

// Load reads and validates the environment.
func Load() (Config, error) {
	core, err := troncenv.LoadCore()
	if err != nil {
		return Config{}, err
	}
	sealer, err := portepg.ParseTokenSealer(troncenv.String("PORTE_TOKEN_KEY", ""))
	if err != nil {
		return Config{}, fmt.Errorf("PORTE_TOKEN_KEY is set but invalid: %w", err)
	}
	summary, err := troncenv.Bool("SONAR_AI_SUMMARY", false)
	if err != nil {
		return Config{}, err
	}
	engines, err := loadEngines()
	if err != nil {
		return Config{}, err
	}
	cfg := Config{
		Engines:     engines,
		Core:        core,
		PublicURL:   strings.TrimSuffix(troncenv.String("SONAR_PUBLIC_URL", "http://localhost:4000"), "/"),
		Sealer:      sealer,
		SuggestURL:  suggestURL(troncenv.String("SONAR_SUGGEST_URL", "")),
		ImageKey:    []byte(troncenv.String("SONAR_IMAGE_KEY", "")),
		Summary:     summary,
		BulleURL:    troncenv.String("BULLE_URL", ""),
		BulleKey:    troncenv.String("SONAR_BULLE_KEY", ""),
		MCPResource: troncenv.String("MCP_RESOURCE", ""),
	}
	return cfg, cfg.loadOIDC()
}

func engineKeys() map[string]string {
	keys := map[string]string{}
	for _, name := range instanceEngines {
		if key := troncenv.String("SONAR_"+strings.ToUpper(name)+"_KEY", ""); key != "" {
			keys[name] = key
		}
	}
	return keys
}

// loadEngines reads the keys and the hosted engine policy. SONAR_WEB_ENGINES
// is the web chain, brave then tavily by default, or "all" for every engine
// at once. SONAR_SUMMARY_ENGINES is where summary sources come from, never
// brave. SONAR_WEB_ENOUGH is how many results end a chain early.
func loadEngines() (Engines, error) {
	enough, err := troncenv.Int("SONAR_WEB_ENOUGH", 5)
	if err != nil {
		return Engines{}, err
	}
	web := search.Chain{Engines: listOr("SONAR_WEB_ENGINES", "brave", "tavily"), Enough: enough}
	if len(web.Engines) == 1 && web.Engines[0] == "all" {
		web.Engines = nil
	}
	sources := search.Chain{Engines: listOr("SONAR_SUMMARY_ENGINES", "tavily", "exa", "parallel"), Enough: enough}
	return Engines{Keys: engineKeys(), Searxng: troncenv.List("SONAR_SEARXNG"), Web: web,
		Sources: sources.Without("brave")}, nil
}

func listOr(key string, fallback ...string) []string {
	if list := troncenv.List(key); len(list) > 0 {
		return list
	}
	return fallback
}

func suggestURL(value string) string {
	switch value {
	case "":
		return "https://duckduckgo.com/ac/"
	case "off":
		return ""
	}
	return value
}

func (c *Config) loadOIDC() error {
	c.OIDC = porte.Config{SSOOnly: true, AppName: "Sonar"}
	issuer := troncenv.String("OIDC_ISSUER", "")
	if issuer == "" {
		return nil
	}
	c.OIDC.Issuer = issuer
	c.OIDC.ClientID = troncenv.String("OIDC_CLIENT_ID", "")
	c.ClientSecret = troncenv.String("OIDC_CLIENT_SECRET", "")
	c.OIDC.ClientSecret = c.ClientSecret
	c.OIDC.RedirectURL = troncenv.String("OIDC_REDIRECT_URL", c.PublicURL+"/api"+porte.RouteCallback)
	c.OIDC.SuccessURL = troncenv.String("OIDC_SUCCESS_URL", c.PublicURL+"/")
	c.OIDC.FailureURL = c.OIDC.SuccessURL
	if c.OIDC.ClientID == "" || c.ClientSecret == "" {
		return fmt.Errorf("OIDC_CLIENT_ID and OIDC_CLIENT_SECRET are required when OIDC_ISSUER is set")
	}
	return nil
}
