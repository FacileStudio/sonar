package search

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/FacileStudio/sonar/websearch"
)

const stageTimeout = 6 * time.Second

var errEmptyChain = errors.New("search: no engine in the chain is enabled")

// Chain is the hosted web policy on sonar's own keys: engines run one at a
// time, in order, and the walk stops once Enough distinct results are in.
// An engine that errors or times out passes the turn to the next one. An
// empty chain means every enabled engine at once.
type Chain struct {
	Engines []string
	Enough  int
}

// Within keeps the chain's engines that are enabled, in order.
func (c Chain) Within(enabled []string) Chain {
	on := map[string]bool{}
	for _, name := range enabled {
		on[name] = true
	}
	kept := Chain{Enough: c.Enough}
	for _, name := range c.Engines {
		if on[name] {
			kept.Engines = append(kept.Engines, name)
		}
	}
	return kept
}

// Without drops one engine from the chain.
func (c Chain) Without(engine string) Chain {
	kept := Chain{Enough: c.Enough}
	for _, name := range c.Engines {
		if name != engine {
			kept.Engines = append(kept.Engines, name)
		}
	}
	return kept
}

func (c Chain) search(ctx context.Context, engines Engines, query string, count int) ([]websearch.Result, error) {
	if len(c.Engines) == 0 {
		return nil, errEmptyChain
	}
	if count <= 0 {
		count = 10
	}
	var out []websearch.Result
	seen := map[string]bool{}
	last := errNoResults
	for i := range c.Engines {
		hits, err := c.stage(ctx, engines, query, count, i)
		if err != nil {
			last = err
			continue
		}
		out = appendNew(out, seen, hits)
		if len(out) >= c.Enough || ctx.Err() != nil {
			break
		}
	}
	if len(out) == 0 {
		return nil, last
	}
	return out[:min(len(out), count)], nil
}

func (c Chain) stage(ctx context.Context, engines Engines, query string, count, i int) ([]websearch.Result, error) {
	if i < len(c.Engines)-1 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, stageTimeout)
		defer cancel()
	}
	return engines.Search(ctx, query, count, []string{c.Engines[i]})
}

func appendNew(out []websearch.Result, seen map[string]bool, hits []websearch.Result) []websearch.Result {
	for _, hit := range hits {
		if key := urlKey(hit.URL); !seen[key] {
			seen[key] = true
			out = append(out, hit)
		}
	}
	return out
}

func urlKey(raw string) string {
	key := strings.ToLower(strings.TrimSpace(raw))
	key = strings.TrimPrefix(strings.TrimPrefix(key, "https://"), "http://")
	return strings.TrimSuffix(strings.TrimPrefix(key, "www."), "/")
}

// SummarySources searches query for a summary on sonar's own keys, through
// the Sources chain only, so a summary never sees results from an engine left
// out of it. It costs one search of the daily quota, refunded on failure.
func (s *Service) SummarySources(ctx context.Context, userID int64, query string, count int) ([]websearch.Result, error) {
	if len(s.Sources.Engines) == 0 {
		return nil, engineError(errEmptyChain)
	}
	if err := s.spend(ctx, userID); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, searchTimeout)
	defer cancel()
	hits, err := s.Sources.search(ctx, s.Instance, query, count)
	if err != nil {
		s.refund(ctx, userID)
	}
	if errors.Is(err, errNoResults) {
		return nil, nil
	}
	return hits, engineError(err)
}
