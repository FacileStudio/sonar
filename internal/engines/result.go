package engines

import (
	"context"
	"errors"
)

// Result is one search hit. Engine records its provenance so the caller can
// show where the result came from and the merge layer can weight by engine.
type Result struct {
	Title    string
	Snippet  string
	URL      string
	Engine   string
	Priority int    `json:"-"`
	FinalURL string `json:"-"`
	Status   int    `json:"-"`
	Cut      bool   `json:"-"`
}

// Sentinel errors classify an engine failure so the dispatcher can choose
// between "fall through to the next engine" (blocked), "back off and retry"
// (transient), and "keep this engine in service".
var (
	ErrBlocked   = errors.New("engine blocked the request (anti-bot/rate wall)")
	ErrTransient = errors.New("transient engine failure")
)

// Engine is a single search backend.
type Engine interface {
	Name() string
	Search(ctx context.Context, query string, count int) ([]Result, error)
}

// Uncacheable is implemented by an engine whose results must never reach a
// cache, such as Brave, whose terms forbid storing results.
type Uncacheable interface {
	Uncached() bool
}

// Cacheable reports whether an engine's results may be stored in a cache.
func Cacheable(e Engine) bool {
	u, ok := e.(Uncacheable)
	return !ok || !u.Uncached()
}
