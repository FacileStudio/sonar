package politeness

import (
	"sync"
	"time"
)

// Breaker is a per-engine circuit breaker. Once an engine trips it (a block
// or repeated transient failure), the breaker stays open for cooldown and the
// dispatcher skips the engine entirely, instead of hammering it in a retry
// loop that converts a transient block into a lasting one.
type Breaker struct {
	cooldown  time.Duration
	openUntil time.Time
	mu        sync.Mutex
}

// NewBreaker returns a circuit breaker that keeps an engine out of service for
// cooldown once it has been tripped.
func NewBreaker(cooldown time.Duration) *Breaker {
	return &Breaker{cooldown: cooldown}
}

// Trip opens the breaker for cooldown from now; a re-trip while open pushes
// the release window forward, so a persistently hostile engine is not
// re-probed every second.
func (b *Breaker) Trip() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.openUntil = time.Now().Add(b.cooldown)
}

// Open reports whether the breaker is currently tripped (engine should be
// skipped) and would not return a stale false immediately after a reopen.
func (b *Breaker) Open() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return time.Now().Before(b.openUntil)
}
