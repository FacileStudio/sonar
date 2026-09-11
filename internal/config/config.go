package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// EngineDef configures one search backend: whether it is enabled, its API
// key (from config or $SONAR_<NAME>_KEY), its base URL where relevant, its
// dispatch priority, and any per-engine result count.
type EngineDef struct {
	Enabled  bool   `yaml:"enabled"`
	APIKey   string `yaml:"apiKey"`
	BaseURL  string `yaml:"baseURL"`
	Zone     string `yaml:"zone"`
	Priority int    `yaml:"priority"`
	Count    int    `yaml:"count"`
}

// Config holds the engine map and the block-minimizing tuning knobs.
type Config struct {
	Engines         map[string]*EngineDef `yaml:"engines"`
	DefaultCount    int                   `yaml:"count"`
	TimeoutSeconds  int                   `yaml:"timeout"`
	MinInterval     time.Duration         `yaml:"minInterval"`
	Jitter          time.Duration         `yaml:"jitter"`
	BreakerCooldown time.Duration         `yaml:"breakerCooldown"`
	CacheDir        string                `yaml:"cacheDir"`
	CacheTTL        time.Duration         `yaml:"cacheTTL"`
}

// Default returns a config with sensible defaults: furet enabled, the scrape
// engines off (ruche's datacenter IP gets them blocked), and gentle pacing.
func Default() *Config {
	return &Config{
		DefaultCount:    10,
		TimeoutSeconds:  20,
		MinInterval:     500 * time.Millisecond,
		Jitter:          150 * time.Millisecond,
		BreakerCooldown: 2 * time.Minute,
		CacheDir:        "~/.cache/sonar",
		CacheTTL:        15 * time.Minute,
		Engines: map[string]*EngineDef{
			"brave":       {Enabled: false, Priority: 1, Count: 10},
			"tavily":      {Enabled: false, Priority: 2, Count: 10},
			"exa":         {Enabled: false, Priority: 3, Count: 10},
			"firecrawl":   {Enabled: false, Priority: 4, Count: 10},
			"serpapi":     {Enabled: false, Priority: 5, Count: 10},
			"browserbase": {Enabled: false, Priority: 6, Count: 10},
			"brightdata":  {Enabled: false, Priority: 7, Count: 10},
			"furet":       {Enabled: true, Priority: 8, BaseURL: "https://furet.facile.studio"},
			"bing":        {Enabled: false, Priority: 9, Count: 10},
			"ddg":         {Enabled: false, Priority: 10, Count: 10},
		},
	}
}

// Load reads the config file at path, merging it over the defaults. A missing
// file is not an error: the defaults stand. Environment overrides are applied
// for any engine that has no key in the file.
func Load(path string) (*Config, error) {
	cfg := Default()
	if path == "" {
		return cfg, nil
	}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}
	if err := yaml.Unmarshal(b, cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	expandTilde(cfg)
	applyEnv(cfg)
	return cfg, nil
}

// expandTilde turns a leading "~/" in CacheDir into the user's home directory,
// since the config defaults to ~/.cache/sonar.
func expandTilde(cfg *Config) {
	if !strings.HasPrefix(cfg.CacheDir, "~/") {
		return
	}
	if home, err := os.UserHomeDir(); err == nil {
		cfg.CacheDir = filepath.Join(home, cfg.CacheDir[2:])
	}
}

func applyEnv(cfg *Config) {
	for name, e := range cfg.Engines {
		if e.APIKey != "" {
			continue
		}
		for _, env := range engineEnvCandidates(name) {
			if key := os.Getenv(env); key != "" {
				e.APIKey = key
				break
			}
		}
	}
	if bd := cfg.Engines["brightdata"]; bd != nil && bd.Zone == "" {
		if z := os.Getenv("BRIGHTDATA_ZONE"); z != "" {
			bd.Zone = z
		}
	}
}

// engineEnvCandidates lists the env vars a key may live in: the canonical
// SONAR_<NAME>_KEY plus the common provider-generic names, so keys stored in
// tiroir under their bare names (EXA_KEY, FIRECRAWL_KEY, ...) are picked up
// without renaming.
func engineEnvCandidates(name string) []string {
	canonical := "SONAR_" + strings.ToUpper(name) + "_KEY"
	switch name {
	case "exa":
		return []string{canonical, "EXA_KEY"}
	case "firecrawl":
		return []string{canonical, "FIRECRAWL_KEY"}
	case "tavily":
		return []string{canonical, "TAVILY_KEY"}
	case "serpapi":
		return []string{canonical, "SERPAPI_API_KEY", "SERPAPI_KEY"}
	case "brave":
		return []string{canonical, "BRAVE_KEY"}
	case "browserbase":
		return []string{canonical, "BROWSERBASE_KEY", "BROWSERBASE_API_KEY"}
	case "brightdata":
		return []string{canonical, "BRIGHTDATA_KEY", "BRIGHTDATA_API_KEY"}
	default:
		return []string{canonical}
	}
}
