package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/FacileStudio/sonar/internal/tiroir"
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

// Config holds the engine map, the user's SearXNG instances, and the
// block-minimizing tuning knobs.
type Config struct {
	Engines         map[string]*EngineDef `yaml:"engines"`
	Searxng         []string              `yaml:"searxng"`
	SearxngPriority int                   `yaml:"searxngPriority"`
	DefaultCount    int                   `yaml:"count"`
	TimeoutSeconds  int                   `yaml:"timeout"`
	MinInterval     time.Duration         `yaml:"minInterval"`
	Jitter          time.Duration         `yaml:"jitter"`
	BreakerCooldown time.Duration         `yaml:"breakerCooldown"`
	MaxConcurrency  int                   `yaml:"maxConcurrency"`
	RetryBackoff    time.Duration         `yaml:"retryBackoff"`
	CacheDir        string                `yaml:"cacheDir"`
	CacheTTL        time.Duration         `yaml:"cacheTTL"`
}

// Default returns a config with sensible defaults: the keyed APIs and scrape
// engines off (ruche's datacenter IP gets them blocked), no SearXNG instance
// (none ships by default), and gentle pacing.
func Default() *Config {
	return &Config{
		DefaultCount:    10,
		TimeoutSeconds:  20,
		MinInterval:     500 * time.Millisecond,
		Jitter:          150 * time.Millisecond,
		BreakerCooldown: 2 * time.Minute,
		MaxConcurrency:  8,
		RetryBackoff:    800 * time.Millisecond,
		CacheDir:        "~/.cache/sonar",
		CacheTTL:        15 * time.Minute,
		SearxngPriority: 9,
		Engines: map[string]*EngineDef{
			"brave":       {Enabled: false, Priority: 1, Count: 10},
			"tavily":      {Enabled: false, Priority: 2, Count: 10},
			"exa":         {Enabled: false, Priority: 3, Count: 10},
			"firecrawl":   {Enabled: false, Priority: 4, Count: 10},
			"serpapi":     {Enabled: false, Priority: 5, Count: 10},
			"browserbase": {Enabled: false, Priority: 6, Count: 10},
			"brightdata":  {Enabled: false, Priority: 7, Count: 10},
			"linkup":      {Enabled: false, Priority: 8, Count: 10},
			"bing":        {Enabled: false, Priority: 10, Count: 10},
			"ddg":         {Enabled: false, Priority: 11, Count: 10},
			"rod":         {Enabled: false, Priority: 12, Count: 1},
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
	if err == nil {
		if perr := yaml.Unmarshal(b, cfg); perr != nil {
			return nil, fmt.Errorf("parse %s: %w", path, perr)
		}
		expandTilde(cfg)
		applyEnv(cfg)
		return cfg, nil
	}
	if os.IsNotExist(err) {
		return cfg, nil
	}
	return nil, err
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
		fillKeyFromEnv(name, e)
		fillKeyFromTiroir(name, e)
	}
	fillBrightdataZone(cfg)
}

// fillKeyFromTiroir fills an engine's key from tiroir when neither the config
// file nor the environment carries one. An MCP-spawned process inherits no
// shell exports, so without this a bare launch would silently skip every keyed
// engine. Only the canonical SONAR_<NAME>_KEY name is consulted, so the sonar
// keys stay namespaced and never collide with generic provider keys. tiroir is
// reached only when the key is still empty, and a missing tiroir (or a key
// absent from it) leaves the key empty without error.
func fillKeyFromTiroir(name string, e *EngineDef) {
	if e.APIKey != "" {
		return
	}
	if v := tiroir.Get(tiroir.CanonicalKey(name)); v != "" {
		e.APIKey = v
	}
}

// fillKeyFromEnv sets an engine's empty key from the first env var that names
// a candidate for it. Returns once a key is set so a later candidate never
// overwrites an earlier one.
func fillKeyFromEnv(name string, e *EngineDef) {
	if e.APIKey != "" {
		return
	}
	for _, env := range tiroir.EnvCandidates(name) {
		if key := os.Getenv(env); key != "" {
			e.APIKey = key
			break
		}
	}
}

// fillBrightdataZone fills the Bright Data SERP zone from its env var when the
// config leaves it empty, so the machine token never has to sit in the file.
// Falls back to tiroir for the same reason the keys do: a bare MCP-spawned
// process inherits no shell exports.
func fillBrightdataZone(cfg *Config) {
	bd := cfg.Engines["brightdata"]
	if bd == nil || bd.Zone != "" {
		return
	}
	if z := os.Getenv("BRIGHTDATA_ZONE"); z != "" {
		bd.Zone = z
		return
	}
	if z := tiroir.Get("SONAR_BRIGHTDATA_ZONE"); z != "" {
		bd.Zone = z
	}
}
