// Package server is `sonar serve`: the hosted sonar at sonar.facile.studio,
// with porte sign-in against Registre, the JSON search API, the image proxy,
// per-user daily quotas from plan entitlements and the SvelteKit client.
//
// The router keeps the root free for /mcp, which a later release mounts next
// to /api with the same search service.
package server

import (
	"fmt"
	"strings"

	"github.com/FacileStudio/porte"
	portepg "github.com/FacileStudio/porte/pg"
	troncenv "github.com/FacileStudio/tronc/env"
)

// instanceEngines are the keyed engines sonar runs on its own keys, each read
// from SONAR_<NAME>_KEY.
var instanceEngines = []string{"brave", "exa", "firecrawl", "linkup", "parallel", "serpapi", "tavily"}

// Config is the process configuration.
type Config struct {
	troncenv.Core
	PublicURL    string
	OIDC         porte.Config
	ClientSecret string
	Sealer       *portepg.TokenSealer
	Keys         map[string]string
	Searxng      []string
	SuggestURL   string
	ImageKey     []byte
	Summary      bool
	BulleURL     string
	BulleKey     string
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
	cfg := Config{
		Core:       core,
		PublicURL:  strings.TrimSuffix(troncenv.String("SONAR_PUBLIC_URL", "http://localhost:4000"), "/"),
		Sealer:     sealer,
		Keys:       engineKeys(),
		Searxng:    troncenv.List("SONAR_SEARXNG"),
		SuggestURL: suggestURL(troncenv.String("SONAR_SUGGEST_URL", "")),
		ImageKey:   []byte(troncenv.String("SONAR_IMAGE_KEY", "")),
		Summary:    summary,
		BulleURL:   troncenv.String("BULLE_URL", ""),
		BulleKey:   troncenv.String("SONAR_BULLE_KEY", ""),
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
