package search

import (
	"context"
	"errors"
	"time"

	"github.com/FacileStudio/sonar/internal/engines"
)

// one honors one unit: served from cache, or fetched through throttle, retry
// and singleflight, caching any fresh results. Engine failures contribute
// nothing; an error is returned only when the run must stop (cancel, throttle).
func (d *Dispatcher) one(ctx context.Context, u *unit, query string, count int) ([]engines.Result, error) {
	if hits, ok := d.cache.Get(u.eng.Name(), query, count); ok {
		return tag(hits, u.prio), nil
	}
	key := u.eng.Name() + "\x00" + query
	v, err, _ := d.flight.Do(key, func() (any, error) {
		return d.fetch(ctx, u, query, count)
	})
	if err != nil {
		return nil, err
	}
	res, _ := v.([]engines.Result)
	return tag(res, u.prio), nil
}

// fetch runs one throttled engine attempt with retry, caching fresh results.
func (d *Dispatcher) fetch(ctx context.Context, u *unit, query string, count int) ([]engines.Result, error) {
	if err := u.throttle.Wait(ctx); err != nil {
		return nil, err
	}
	res, err := d.run(ctx, u, query, count)
	if err == nil {
		if len(res) > 0 {
			d.cache.Put(u.eng.Name(), query, res)
		}
		return res, nil
	}
	tripBlocked(u, err)
	return []engines.Result{}, nil
}

func tag(in []engines.Result, prio int) []engines.Result {
	out := make([]engines.Result, len(in))
	for i, r := range in {
		r.Priority = prio
		out[i] = r
	}
	return out
}

// tripBlocked parks a single engine when the run identifies it as blocked so
// it stays out of service for its cooldown instead of being re-hammered.
func tripBlocked(u *unit, err error) {
	if !errors.Is(err, engines.ErrBlocked) {
		return
	}
	u.breaker.Trip()
}

func (d *Dispatcher) run(ctx context.Context, u *unit, query string, count int) ([]engines.Result, error) {
	res, err := u.eng.Search(ctx, query, count)
	if err != nil && errors.Is(err, engines.ErrTransient) && ctx.Err() == nil {
		t := time.NewTimer(d.retryWait)
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
