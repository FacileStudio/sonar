// Package search wires the engines together with block-minimizing behaviour:
// keyed engines dispatch concurrently, scrape engines stay sequential behind a
// semaphore, each engine is paced by its own politeness throttle, and every
// engine carries a circuit breaker and a result cache.
package search

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/FacileStudio/sonar/internal/cache"
	"github.com/FacileStudio/sonar/internal/config"
	"github.com/FacileStudio/sonar/internal/engines"
	"github.com/FacileStudio/sonar/internal/httpc"
	"github.com/FacileStudio/sonar/internal/merge"
	"github.com/FacileStudio/sonar/internal/politeness"
	"golang.org/x/sync/singleflight"
)

type unit struct {
	name     string
	eng      engines.Engine
	breaker  *politeness.Breaker
	throttle *politeness.Throttle
	prio     int
	scrape   bool
}

// Dispatcher runs every enabled engine with per-engine throttling, circuit
// breakers and a result cache, collecting the results that survive. Keyed
// engines run concurrently against distinct hosts; scrape engines are held to
// one at a time so they never burst against the same IP.
type Dispatcher struct {
	order     []unit
	cache     *cache.Cache
	scrape    chan struct{}
	inflight  chan struct{}
	flight    singleflight.Group
	retryWait time.Duration
	timeout   time.Duration
}

// Query builds a dispatcher and runs one full ranked search under a timeout
// budget. Shared by the search command so a one-shot process gets the whole
// pipeline in one call. Include names a subset of engines to run; empty means
// every enabled one.
func Query(ctx context.Context, cfg *config.Config, query string, count int, include []string) ([]engines.Result, error) {
	d, err := NewDispatcher(cfg)
	if err != nil {
		return nil, fmt.Errorf("engine setup: %w", err)
	}
	return d.Query(ctx, query, count, include)
}

// Query runs one ranked search: search under a timeout budget, then rank and
// cap the survivors. Long-lived callers (the MCP server) reuse the dispatcher
// across queries so breakers and throttle warm-up persist. Include names a
// subset of engines to run; empty means every enabled one.
func (d *Dispatcher) Query(ctx context.Context, query string, count int, include []string) ([]engines.Result, error) {
	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()
	results, err := d.Search(ctx, query, count, include)
	if err != nil {
		return nil, err
	}
	return merge.Rank(results, count), nil
}

// NewDispatcher builds a dispatcher from config, wiring each enabled engine to
// its own throttle and circuit breaker.
func NewDispatcher(cfg *config.Config) (*Dispatcher, error) {
	c, err := cache.New(cfg.CacheDir, cfg.CacheTTL)
	if err != nil {
		return nil, err
	}
	client := httpc.Client(time.Duration(cfg.TimeoutSeconds) * time.Second)
	built := buildEngineUnits(cfg, client)
	sort.SliceStable(built, func(i, j int) bool { return built[i].prio < built[j].prio })
	d := &Dispatcher{
		cache:     c,
		scrape:    make(chan struct{}, 1),
		inflight:  make(chan struct{}, maxConcurrency(cfg)),
		retryWait: retryBackoff(cfg),
		timeout:   time.Duration(cfg.TimeoutSeconds)*time.Second + 5*time.Second,
	}
	for _, b := range built {
		d.order = append(d.order, unit{
			name:     b.name,
			eng:      b.eng,
			breaker:  politeness.NewBreaker(cfg.BreakerCooldown),
			throttle: politeness.NewThrottle(cfg.MinInterval, cfg.Jitter),
			prio:     b.prio,
			scrape:   b.scrape,
		})
	}
	return d, nil
}

func maxConcurrency(cfg *config.Config) int {
	if cfg.MaxConcurrency > 0 {
		return cfg.MaxConcurrency
	}
	return 8
}

func retryBackoff(cfg *config.Config) time.Duration {
	if cfg.RetryBackoff > 0 {
		return cfg.RetryBackoff
	}
	return 800 * time.Millisecond
}

// Search fans out to the selected engines — every enabled one when include is
// empty — falling through on failure and merging the survivors. Returns
// combined results, or an error only when no engine answered with results.
// A name in include that matches no enabled engine is an error naming what is
// actually available.
func (d *Dispatcher) Search(ctx context.Context, query string, count int, include []string) ([]engines.Result, error) {
	units, err := d.selectUnits(include)
	if err != nil {
		return nil, err
	}
	return d.fanOut(ctx, units, query, count)
}
