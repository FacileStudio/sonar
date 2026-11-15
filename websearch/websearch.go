// Package websearch exposes sonar's multi-engine search as a Go library.
//
// Keys come only from Config: the package never reads a config file, an
// environment variable or tiroir, and its result cache lives in memory. A
// caller that holds keys for several tenants builds one Client per key set.
package websearch

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/FacileStudio/sonar/internal/config"
	"github.com/FacileStudio/sonar/internal/engines"
	"github.com/FacileStudio/sonar/internal/search"
)

// ErrNoEngines is returned by Search when the Config enabled no engine.
var ErrNoEngines = errors.New("websearch: no engine is configured")

const defaultCount = 10

// Config selects the engines a Client runs. Keys maps a keyed engine name
// (tavily, parallel, exa, brave, ...) to its API key; an empty key leaves the
// engine off. Searxng lists self-hosted SearXNG base URLs, one engine each. Pool lists
// SearXNG instances, public ones included, that one engine named "searxpool"
// rotates through, at most PoolTries per search (two when zero). Free names
// key-less scrape engines to switch on (ddg, bing), which datacenter IPs
// often get blocked on. Timeout bounds one
// engine request and CacheTTL how long results are reused; zero keeps sonar's
// defaults. Brave results are never cached, whatever CacheTTL says.
type Config struct {
	Keys      map[string]string
	Searxng   []string
	Pool      []string
	PoolTries int
	Free      []string
	Timeout   time.Duration
	CacheTTL  time.Duration
}

// Result is one ranked search hit.
type Result struct {
	Title   string
	URL     string
	Snippet string
	Engine  string
}

// Client runs searches over the engines its Config enabled. It is safe for
// concurrent use, and reusing it keeps circuit breakers, throttles and the
// cache warm across queries.
type Client struct {
	dispatcher *search.Dispatcher
	engines    []string
	brave      *engines.Brave
}

// New builds a Client from cfg. A key for an engine sonar does not know, or
// for one that needs no key, is an error.
func New(cfg Config) (*Client, error) {
	sc, names, err := sonarConfig(cfg)
	if err != nil {
		return nil, err
	}
	d, err := search.NewDispatcher(sc)
	if err != nil {
		return nil, fmt.Errorf("websearch: %w", err)
	}
	return &Client{dispatcher: d, engines: names, brave: newBrave(cfg.Keys["brave"], cfg.Timeout)}, nil
}

// Engines lists the enabled engine names in sorted order.
func (c *Client) Engines() []string {
	return append([]string(nil), c.engines...)
}

// Search runs one ranked query and returns at most count results. A count of
// zero or less means ten. Include restricts the run to the named engines;
// empty runs every enabled one.
func (c *Client) Search(ctx context.Context, query string, count int, include []string) ([]Result, error) {
	if len(c.engines) == 0 {
		return nil, ErrNoEngines
	}
	if count <= 0 {
		count = defaultCount
	}
	hits, err := c.dispatcher.Query(ctx, query, count, include)
	if err != nil {
		return nil, err
	}
	return convert(hits), nil
}

func sonarConfig(cfg Config) (*config.Config, []string, error) {
	sc := config.Default()
	sc.CacheDir = ""
	if cfg.Timeout > 0 {
		sc.TimeoutSeconds = max(1, int(cfg.Timeout/time.Second))
	}
	if cfg.CacheTTL > 0 {
		sc.CacheTTL = cfg.CacheTTL
	}
	names, err := enableKeyed(sc, cfg.Keys)
	if err != nil {
		return nil, nil, err
	}
	free, err := enableFree(sc, cfg.Free)
	if err != nil {
		return nil, nil, err
	}
	names = append(names, free...)
	names = append(names, addSearxng(sc, cfg)...)
	sort.Strings(names)
	return sc, names, nil
}

func enableKeyed(sc *config.Config, keys map[string]string) ([]string, error) {
	var names []string
	for name, key := range keys {
		def, ok := sc.Engines[name]
		if !ok || !search.NeedsKey(name) {
			return nil, fmt.Errorf("websearch: unknown keyed engine %q", name)
		}
		if key == "" {
			continue
		}
		def.Enabled, def.APIKey = true, key
		names = append(names, name)
	}
	return names, nil
}

func addSearxng(sc *config.Config, cfg Config) []string {
	var names []string
	if len(cfg.Pool) > 0 {
		sc.SearxngPool, sc.SearxngPoolTries = append([]string(nil), cfg.Pool...), cfg.PoolTries
		names = append(names, engines.PoolName)
	}
	sc.Searxng = append([]string(nil), cfg.Searxng...)
	for i := range sc.Searxng {
		names = append(names, search.SearxngName(sc.Searxng, i))
	}
	return names
}

func convert(hits []engines.Result) []Result {
	out := make([]Result, 0, len(hits))
	for _, h := range hits {
		out = append(out, Result{Title: h.Title, URL: h.URL, Snippet: h.Snippet, Engine: h.Engine})
	}
	return out
}

func enableFree(sc *config.Config, free []string) ([]string, error) {
	var names []string
	for _, name := range free {
		def, ok := sc.Engines[name]
		if !ok || search.NeedsKey(name) || !search.IsScrape(name) {
			return nil, fmt.Errorf("websearch: %q is not a free scrape engine", name)
		}
		def.Enabled = true
		names = append(names, name)
	}
	return names, nil
}
