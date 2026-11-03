// Package cache stores recent engine results on disk so repeat queries do not
// re-hit the engine. Agents repeat queries often; a hit served from cache is
// one fewer request to an IP already fighting search-engine reputation.
package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/FacileStudio/sonar/internal/engines"
)

type entry struct {
	Expires time.Time        `json:"expires"`
	Results []engines.Result `json:"results"`
}

// Cache stores recent engine results on disk keyed by engine+query. Agents
// repeat queries often; a hit served here is one fewer request to an IP already
// fighting search-engine reputation. A cache built with an empty dir keeps
// entries in memory only, for library callers that must not touch the disk.
type Cache struct {
	mu  sync.Mutex
	dir string
	ttl time.Duration
	mem map[string]entry
}

// New creates a cache directory (0600) and returns a Cache with the given TTL.
// An empty dir returns an in-memory cache that never touches the filesystem.
func New(dir string, ttl time.Duration) (*Cache, error) {
	if dir == "" {
		return &Cache{ttl: ttl, mem: map[string]entry{}}, nil
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	return &Cache{dir: dir, ttl: ttl}, nil
}

// Get returns cached results for the engine+query if present and fresh.
func (c *Cache) Get(engine, query string, count int) ([]engines.Result, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.load(engine, query)
	if !ok || time.Now().After(e.Expires) {
		return nil, false
	}
	if len(e.Results) > count {
		e.Results = e.Results[:count]
	}
	return e.Results, true
}

// Put stores results (keyed by engine+query) with the cache TTL.
func (c *Cache) Put(engine, query string, results []engines.Result) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	e := entry{Expires: time.Now().Add(c.ttl), Results: results}
	if c.mem != nil {
		c.mem[c.key(engine, query)] = e
		return nil
	}
	b, _ := json.Marshal(e)
	return os.WriteFile(c.path(engine, query), b, 0o600)
}

func (c *Cache) load(engine, query string) (entry, bool) {
	if c.mem != nil {
		e, ok := c.mem[c.key(engine, query)]
		return e, ok
	}
	b, err := os.ReadFile(c.path(engine, query))
	if err != nil {
		return entry{}, false
	}
	var e entry
	if json.Unmarshal(b, &e) != nil {
		return entry{}, false
	}
	return e, true
}

func (c *Cache) key(engine, query string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(query)))
	return engine + "_" + hex.EncodeToString(sum[:])
}

func (c *Cache) path(engine, query string) string {
	return filepath.Join(c.dir, c.key(engine, query)+".json")
}
