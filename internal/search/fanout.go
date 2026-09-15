package search

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/FacileStudio/sonar/internal/engines"
)

// selectUnits filters the dispatcher's engine order down to the requested
// names, keeping the configured priority order. Empty means every enabled
// engine; an unknown or disabled name fails with the list of usable ones.
func (d *Dispatcher) selectUnits(include []string) ([]unit, error) {
	if len(include) == 0 {
		return d.order, nil
	}
	byName := make(map[string]*unit, len(d.order))
	available := make([]string, 0, len(d.order))
	for i := range d.order {
		u := &d.order[i]
		byName[u.name] = u
		available = append(available, u.name)
	}
	units := make([]unit, 0, len(include))
	for _, name := range include {
		u, ok := byName[name]
		if !ok {
			return nil, fmt.Errorf("unknown or disabled engine %q (enabled: %s)", name, strings.Join(available, ", "))
		}
		units = append(units, *u)
	}
	return units, nil
}

// fanOut runs the selected units, collecting the results that survive. Keyed
// engines run concurrently; scrapers hold the single scrape slot.
func (d *Dispatcher) fanOut(ctx context.Context, units []unit, query string, count int) ([]engines.Result, error) {
	var (
		mu  sync.Mutex
		all []engines.Result
		wg  sync.WaitGroup
	)
	for i := range units {
		u := &units[i]
		if u.breaker.Open() || !d.acquire(ctx, u) {
			continue
		}
		wg.Go(func() {
			defer d.release(u)
			if res, err := d.one(ctx, u, query, count); err == nil {
				mu.Lock()
				all = append(all, res...)
				mu.Unlock()
			}
		})
	}
	wg.Wait()
	if len(all) == 0 {
		return nil, errors.New("no engine returned results")
	}
	return all, nil
}

// acquire takes the concurrency slots a unit needs before its goroutine
// starts: the global cap for every engine, plus the single scrape slot for the
// scrapers, so scrapes never overlap each other.
func (d *Dispatcher) acquire(ctx context.Context, u *unit) bool {
	slots := []chan struct{}{d.inflight}
	if u.scrape {
		slots = []chan struct{}{d.scrape, d.inflight}
	}
	for _, s := range slots {
		select {
		case s <- struct{}{}:
		case <-ctx.Done():
			return false
		}
	}
	return true
}

func (d *Dispatcher) release(u *unit) {
	<-d.inflight
	if u.scrape {
		<-d.scrape
	}
}
