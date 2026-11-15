package engines

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/FacileStudio/sonar/internal/politeness"
)

// PoolName is the engine name a SearXNG pool answers to in the web chain.
const PoolName = "searxpool"

const defaultPoolTries = 2

// PoolOptions tunes a SearxPool. Zero values pick conservative defaults: two
// instances tried per search, a 3 second attempt, a two minute cooldown after
// a failure and no spacing between calls to one instance.
type PoolOptions struct {
	MaxTries       int
	AttemptTimeout time.Duration
	Cooldown       time.Duration
	MinInterval    time.Duration
	Jitter         time.Duration
}

type poolMember struct {
	base     string
	breaker  *politeness.Breaker
	throttle *politeness.Throttle
}

// SearxPool searches a list of SearXNG instances, self-hosted or public, one
// at a time. Each search starts at the next instance in rotation and tries at
// most MaxTries of them. An instance that blocks, rate-limits, answers with a
// page instead of JSON or fails is parked by its own circuit breaker. The
// query travels only in a POST body, and no error carries it or a request URL.
type SearxPool struct {
	Client  *http.Client
	Count   int
	opts    PoolOptions
	members []*poolMember
	turn    atomic.Uint32
}

// NewSearxPool builds a pool over bases, dropping blank entries and trailing
// slashes. A client without a timeout still gets one per attempt.
func NewSearxPool(bases []string, client *http.Client, opts PoolOptions) *SearxPool {
	if opts.MaxTries <= 0 {
		opts.MaxTries = defaultPoolTries
	}
	if opts.AttemptTimeout <= 0 {
		opts.AttemptTimeout = 3 * time.Second
	}
	if opts.Cooldown <= 0 {
		opts.Cooldown = 2 * time.Minute
	}
	p := &SearxPool{Client: client, opts: opts}
	for _, base := range bases {
		if base = strings.TrimRight(strings.TrimSpace(base), "/"); base != "" {
			p.members = append(p.members, &poolMember{base: base,
				breaker:  politeness.NewBreaker(opts.Cooldown),
				throttle: politeness.NewThrottle(opts.MinInterval, opts.Jitter)})
		}
	}
	return p
}

func (p *SearxPool) Name() string { return PoolName }

// Size is the number of instances in the pool.
func (p *SearxPool) Size() int { return len(p.members) }

// Search returns the first instance's non-empty answer. It reports ErrBlocked
// when every instance it tried refused the request, and ErrTransient when
// some failed another way or none was available.
func (p *SearxPool) Search(ctx context.Context, query string, count int) ([]Result, error) {
	start := int(p.turn.Add(1) - 1)
	var tally poolTally
	for i := 0; i < len(p.members) && tally.tried < p.opts.MaxTries && ctx.Err() == nil; i++ {
		m := p.members[(start+i)%len(p.members)]
		if m.breaker.Open() {
			continue
		}
		hits, err := p.attempt(ctx, m, query, count)
		if err == nil && len(hits) > 0 {
			return hits, nil
		}
		tally.record(err)
	}
	return nil, tally.failure()
}

type poolTally struct{ tried, blocked int }

func (t *poolTally) record(err error) {
	t.tried++
	if errors.Is(err, ErrBlocked) {
		t.blocked++
	}
}

func (t *poolTally) failure() error {
	switch {
	case t.tried == 0:
		return fmt.Errorf("%w: every instance is cooling down", ErrTransient)
	case t.blocked == t.tried:
		return ErrBlocked
	}
	return fmt.Errorf("%w: no instance answered", ErrTransient)
}

func (p *SearxPool) attempt(ctx context.Context, m *poolMember, query string, count int) ([]Result, error) {
	if err := m.throttle.Wait(ctx); err != nil {
		return nil, fmt.Errorf("%w: wait cancelled", ErrTransient)
	}
	ctx, cancel := context.WithTimeout(ctx, p.opts.AttemptTimeout)
	defer cancel()
	hits, err := p.call(ctx, m.base, query)
	if err != nil {
		m.breaker.Trip()
		return nil, err
	}
	limit := count
	if p.Count > 0 && (limit <= 0 || p.Count < limit) {
		limit = p.Count
	}
	if limit > 0 && len(hits) > limit {
		hits = hits[:limit]
	}
	return hits, nil
}
