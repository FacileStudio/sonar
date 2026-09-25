package search

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

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

func (d *Dispatcher) launchUnit(ctx context.Context, u *unit, query string, count int, ch chan<- []engines.Result) {
	if !d.acquire(ctx, u) {
		ch <- nil
		return
	}
	defer d.release(u)
	res, err := d.one(ctx, u, query, count)
	if err == nil && len(res) > 0 {
		ch <- res
		return
	}
	ch <- nil
}

func collectResults(ctx context.Context, cancel context.CancelFunc, ch <-chan []engines.Result, count, total int) []engines.Result {
	var (
		all     []engines.Result
		timerCh <-chan time.Time
	)
	for range total {
		select {
		case res := <-ch:
			all = append(all, res...)
			if len(all) >= count && timerCh == nil {
				t := time.NewTimer(1200 * time.Millisecond)
				defer t.Stop()
				timerCh = t.C
			}
		case <-timerCh:
			cancel()
			return all
		case <-ctx.Done():
			return all
		}
	}
	return all
}

// fanOut runs the selected units, collecting the results that survive. Keyed
// engines run concurrently; scrapers hold the single scrape slot.
func (d *Dispatcher) fanOut(ctx context.Context, units []unit, query string, count int) ([]engines.Result, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	resCh := make(chan []engines.Result, len(units))
	var launched int
	for i := range units {
		u := &units[i]
		if u.breaker.Open() {
			continue
		}
		launched++
		go d.launchUnit(ctx, u, query, count, resCh)
	}
	if launched == 0 {
		return nil, errors.New("no engine returned results")
	}
	all := collectResults(ctx, cancel, resCh, count, launched)
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
