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
// fighting search-engine reputation.
type Cache struct {
	mu  sync.Mutex
	dir string
	ttl time.Duration
}

// New creates a cache directory (0600) and returns a Cache with the given TTL.
func New(dir string, ttl time.Duration) (*Cache, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	return &Cache{dir: dir, ttl: ttl}, nil
}

// Get returns cached results for the engine+query if present and fresh.
func (c *Cache) Get(engine, query string, count int) ([]engines.Result, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	b, err := os.ReadFile(c.path(engine, query))
	if err != nil {
		return nil, false
	}
	var e entry
	if json.Unmarshal(b, &e) != nil || time.Now().After(e.Expires) {
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
	b, _ := json.Marshal(e)
	return os.WriteFile(c.path(engine, query), b, 0o600)
}

func (c *Cache) path(engine, query string) string {
	sum := sha256.Sum256([]byte(strings.ToLower(query)))
	return filepath.Join(c.dir, engine+"_"+hex.EncodeToString(sum[:])+".json")
}
