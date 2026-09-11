// Package politeness keeps calls to block-prone engines gentle and spaced out.
package politeness

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

// Throttle enforces a minimum interval between calls to one engine, with a
// random jitter added so calls do not cluster into a detectable burst.
type Throttle struct {
	minInterval time.Duration
	jitter      time.Duration
	last        time.Time
	mu          sync.Mutex
}

// NewThrottle returns a throttle that spaces calls to one engine by
// minInterval plus a jitter of up to jitter.
func NewThrottle(minInterval, jitter time.Duration) *Throttle {
	return &Throttle{minInterval: minInterval, jitter: jitter}
}

// Wait blocks until the minimum interval since the last call has elapsed,
// plus up to jitter. Returns ctx.Err if the context is cancelled while pacing.
func (t *Throttle) Wait(ctx context.Context) error {
	t.mu.Lock()
	delay := time.Until(t.last.Add(t.minInterval))
	t.last = time.Now()
	if t.jitter > 0 {
		delay += time.Duration(rand.Int63n(int64(t.jitter)))
	}
	t.mu.Unlock()
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
