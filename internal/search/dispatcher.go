// Package search wires the engines together with block-minimizing behaviour:
// sequential dispatch (never a simultaneous multi-engine burst), a politeness
// throttle, per-engine circuit breakers, and a result cache.
package search

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/FacileStudio/sonar/internal/cache"
	"github.com/FacileStudio/sonar/internal/config"
	"github.com/FacileStudio/sonar/internal/engines"
	"github.com/FacileStudio/sonar/internal/httpc"
	"github.com/FacileStudio/sonar/internal/merge"
	"github.com/FacileStudio/sonar/internal/politeness"
)

type unit struct {
	eng     engines.Engine
	breaker *politeness.Breaker
}

// Dispatcher runs every enabled engine in priority order with throttling,
// circuit breakers and a result cache, collecting the results that survive.
type Dispatcher struct {
	order    []unit
	cache    *cache.Cache
	throttle *politeness.Throttle
}

// Query runs one full ranked search for a user-facing query: build the
// dispatcher from config, search under the doubled timeout, then rank and cap
// the survivors. Shared by the search command and the MCP server so both
// surfaces return identical results for the same query.
func Query(ctx context.Context, cfg *config.Config, query string, count int) ([]engines.Result, error) {
	d, err := NewDispatcher(cfg)
	if err != nil {
		return nil, fmt.Errorf("engine setup: %w", err)
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(cfg.TimeoutSeconds)*2*time.Second)
	defer cancel()
	results, err := d.Search(ctx, query, count)
	if err != nil {
		return nil, err
	}
	return merge.Rank(results, count), nil
}

// NewDispatcher builds a dispatcher from config, wiring each enabled engine
// to its own circuit breaker behind a shared politeness throttle.
func NewDispatcher(cfg *config.Config) (*Dispatcher, error) {
	throttle := politeness.NewThrottle(cfg.MinInterval, cfg.Jitter)
	c, err := cache.New(cfg.CacheDir, cfg.CacheTTL)
	if err != nil {
		return nil, err
	}
	d := &Dispatcher{cache: c, throttle: throttle}
	client := httpc.Client(time.Duration(cfg.TimeoutSeconds) * time.Second)
	built := buildEngineUnits(cfg, client)
	sort.SliceStable(built, func(i, j int) bool { return built[i].prio < built[j].prio })
	for _, b := range built {
		d.order = append(d.order, unit{eng: b.eng, breaker: politeness.NewBreaker(cfg.BreakerCooldown)})
	}
	return d, nil
}

// Search fans out to every enabled engine in priority order, falling through
// on failure and merging the survivors. Returns combined results, or an error
// only when no engine answered with results.
func (d *Dispatcher) Search(ctx context.Context, query string, count int) ([]engines.Result, error) {
	var all []engines.Result
	for _, u := range d.order {
		if ctx.Err() != nil {
			return all, ctx.Err()
		}
		if u.breaker.Open() {
			continue
		}
		if hits, ok := d.cache.Get(u.eng.Name(), query, count); ok {
			all = append(all, hits...)
			continue
		}
		if err := d.throttle.Wait(ctx); err != nil {
			return all, err
		}
		res, err := u.run(ctx, query, count)
		if err != nil {
			if errors.Is(err, engines.ErrBlocked) {
				u.breaker.Trip()
			}
			continue
		}
		if len(res) > 0 {
			d.cache.Put(u.eng.Name(), query, res)
		}
		all = append(all, res...)
	}
	if len(all) == 0 {
		return nil, errors.New("no engine returned results")
	}
	return all, nil
}

func (u *unit) run(ctx context.Context, query string, count int) ([]engines.Result, error) {
	res, err := u.eng.Search(ctx, query, count)
	if err != nil && errors.Is(err, engines.ErrTransient) && ctx.Err() == nil {
		t := time.NewTimer(800 * time.Millisecond)
		defer t.Stop()
		select {
		case <-t.C:
			return u.eng.Search(ctx, query, count)
		case <-ctx.Done():
			return res, err
		}
	}
	return res, err
}
