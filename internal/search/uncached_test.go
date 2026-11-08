package search

import (
	"context"
	"testing"
	"time"

	"github.com/FacileStudio/sonar/internal/politeness"
)

type uncachedStub struct {
	stubEngine
}

func (u *uncachedStub) Uncached() bool { return true }

func TestUncachedEngineIsQueriedEveryTime(t *testing.T) {
	brave := &uncachedStub{stubEngine{name: "brave", hits: hitsFor("brave", "", 2)}}
	tavily := &stubEngine{name: "tavily", hits: hitsFor("tavily", "", 2)}
	d := testDispatcher(t,
		unit{eng: brave, breaker: politeness.NewBreaker(time.Minute), throttle: politeness.NewThrottle(0, 0), prio: 1},
		unit{eng: tavily, breaker: politeness.NewBreaker(time.Minute), throttle: politeness.NewThrottle(0, 0), prio: 2},
	)
	for range 3 {
		if _, err := d.Search(context.Background(), "same query", 4, nil); err != nil {
			t.Fatalf("search: %v", err)
		}
	}
	if brave.calls != 3 {
		t.Fatalf("brave called %d times, want 3: its results must never be cached", brave.calls)
	}
	if tavily.calls != 1 {
		t.Fatalf("tavily called %d times, want 1: its results are cacheable", tavily.calls)
	}
	if _, ok := d.cache.Get("brave", "same query", 4); ok {
		t.Fatal("brave results reached the cache")
	}
}
