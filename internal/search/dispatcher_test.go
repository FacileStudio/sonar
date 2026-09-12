package search

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/FacileStudio/sonar/internal/cache"
	"github.com/FacileStudio/sonar/internal/engines"
	"github.com/FacileStudio/sonar/internal/merge"
	"github.com/FacileStudio/sonar/internal/politeness"
)

type stubEngine struct {
	name  string
	prio  int
	hits  []engines.Result
	err   error
	delay time.Duration
	calls int
	mu    sync.Mutex
}

func (s *stubEngine) Name() string { return s.name }

func (s *stubEngine) Search(ctx context.Context, query string, count int) ([]engines.Result, error) {
	s.mu.Lock()
	s.calls++
	s.mu.Unlock()
	if s.delay > 0 {
		select {
		case <-time.After(s.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	if s.err != nil {
		return nil, s.err
	}
	out := make([]engines.Result, len(s.hits))
	copy(out, s.hits)
	return out, nil
}

func testDispatcher(t *testing.T, units ...unit) *Dispatcher {
	t.Helper()
	dir := t.TempDir()
	c, err := cache.New(dir, time.Minute)
	if err != nil {
		t.Fatalf("cache: %v", err)
	}
	d := &Dispatcher{
		cache:     c,
		scrape:    make(chan struct{}, 1),
		inflight:  make(chan struct{}, 8),
		retryWait: time.Millisecond,
		timeout:   5 * time.Second,
	}
	d.order = append(d.order, units...)
	return d
}

func hitsFor(engine, prefix string, n int) []engines.Result {
	out := make([]engines.Result, 0, n)
	for i := range n {
		out = append(out, engines.Result{
			Title:  fmt.Sprintf("%s %d", engine, i),
			URL:    fmt.Sprintf("https://%s.example/%s%d", engine, prefix, i),
			Engine: engine,
		})
	}
	return out
}

func TestSearchRunsEnginesConcurrently(t *testing.T) {
	slow := &stubEngine{name: "slow", hits: hitsFor("slow", "", 3), delay: 200 * time.Millisecond}
	fast := &stubEngine{name: "fast", hits: hitsFor("fast", "", 3)}
	d := testDispatcher(t,
		unit{eng: slow, breaker: politeness.NewBreaker(time.Minute), throttle: politeness.NewThrottle(0, 0), prio: 1},
		unit{eng: fast, breaker: politeness.NewBreaker(time.Minute), throttle: politeness.NewThrottle(0, 0), prio: 2},
	)
	start := time.Now()
	res, err := d.Search(context.Background(), "q", 10, nil)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if elapsed := time.Since(start); elapsed >= 400*time.Millisecond {
		t.Fatalf("search took %v, engines did not run concurrently", elapsed)
	}
	if len(res) != 6 {
		t.Fatalf("got %d results, want 6", len(res))
	}
}

func TestScrapeEnginesStaySequential(t *testing.T) {
	bing := &stubEngine{name: "bing", hits: hitsFor("bing", "", 1), delay: 60 * time.Millisecond}
	ddg := &stubEngine{name: "ddg", hits: hitsFor("ddg", "", 1), delay: 60 * time.Millisecond}
	d := testDispatcher(t,
		unit{eng: bing, breaker: politeness.NewBreaker(time.Minute), throttle: politeness.NewThrottle(0, 0), prio: 1, scrape: true},
		unit{eng: ddg, breaker: politeness.NewBreaker(time.Minute), throttle: politeness.NewThrottle(0, 0), prio: 2, scrape: true},
	)
	start := time.Now()
	if _, err := d.Search(context.Background(), "q", 10, nil); err != nil {
		t.Fatalf("search: %v", err)
	}
	if elapsed := time.Since(start); elapsed < 120*time.Millisecond {
		t.Fatalf("scrape engines overlapped: %v for two 60ms calls", elapsed)
	}
}

func TestSearchRankingMatchesSerialDispatch(t *testing.T) {
	brave := &stubEngine{name: "brave", hits: hitsFor("brave", "", 4)}
	bing := &stubEngine{name: "bing", hits: hitsFor("bing", "shared", 3)}
	ddg := &stubEngine{name: "ddg", hits: hitsFor("ddg", "shared", 3)}
	units := []unit{
		{eng: brave, breaker: politeness.NewBreaker(time.Minute), throttle: politeness.NewThrottle(0, 0), prio: 1},
		{eng: bing, breaker: politeness.NewBreaker(time.Minute), throttle: politeness.NewThrottle(0, 0), prio: 10, scrape: true},
		{eng: ddg, breaker: politeness.NewBreaker(time.Minute), throttle: politeness.NewThrottle(0, 0), prio: 11, scrape: true},
	}
	d := testDispatcher(t, units...)
	res, err := d.Search(context.Background(), "q", 10, nil)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	ranked := merge.Rank(res, 10)
	var want []engines.Result
	for _, s := range []*stubEngine{brave, bing, ddg} {
		want = append(want, s.hits...)
	}
	got := merge.Rank(want, 10)
	if len(ranked) != len(got) {
		t.Fatalf("ranked %d results, want %d", len(ranked), len(got))
	}
	for i := range ranked {
		if ranked[i].URL != got[i].URL {
			t.Fatalf("position %d = %s, want %s", i, ranked[i].URL, got[i].URL)
		}
	}
}

func TestFailingEngineFallsThrough(t *testing.T) {
	dead := &stubEngine{name: "dead", err: errors.New("boom")}
	live := &stubEngine{name: "brave", hits: hitsFor("brave", "", 2)}
	d := testDispatcher(t,
		unit{eng: dead, breaker: politeness.NewBreaker(time.Minute), throttle: politeness.NewThrottle(0, 0), prio: 1},
		unit{eng: live, breaker: politeness.NewBreaker(time.Minute), throttle: politeness.NewThrottle(0, 0), prio: 2},
	)
	res, err := d.Search(context.Background(), "q", 10, nil)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(res) != 2 {
		t.Fatalf("got %d results, want the live engine's 2", len(res))
	}
}

func TestCancelledContextStopsDispatch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	d := testDispatcher(t, unit{eng: &stubEngine{name: "brave"}, breaker: politeness.NewBreaker(time.Minute), throttle: politeness.NewThrottle(0, 0), prio: 1})
	if _, err := d.Search(ctx, "q", 10, nil); !errors.Is(err, context.Canceled) && err.Error() != "no engine returned results" {
		t.Fatalf("err = %v, want cancel or empty result", err)
	}
}

func TestSearchSelectsEngines(t *testing.T) {
	brave := &stubEngine{name: "brave", hits: hitsFor("brave", "", 2)}
	bing := &stubEngine{name: "bing", hits: hitsFor("bing", "", 2)}
	d := testDispatcher(t,
		unit{name: "brave", eng: brave, breaker: politeness.NewBreaker(time.Minute), throttle: politeness.NewThrottle(0, 0), prio: 1},
		unit{name: "bing", eng: bing, breaker: politeness.NewBreaker(time.Minute), throttle: politeness.NewThrottle(0, 0), prio: 2, scrape: true},
	)
	res, err := d.Search(context.Background(), "q", 10, []string{"bing"})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	for _, r := range res {
		if r.Engine != "bing" {
			t.Fatalf("got result from %q, want only bing", r.Engine)
		}
	}
	if _, err := d.Search(context.Background(), "q", 10, []string{"exa"}); err == nil {
		t.Fatal("unknown engine accepted")
	}
}

func TestSingleflightDistinguishesCounts(t *testing.T) {
	eng := &stubEngine{name: "brave", hits: hitsFor("brave", "", 5)}
	d := testDispatcher(t, unit{eng: eng, breaker: politeness.NewBreaker(time.Minute), throttle: politeness.NewThrottle(0, 0), prio: 1})
	var wg sync.WaitGroup
	for _, count := range []int{3, 7} {
		wg.Go(func() {
			if _, err := d.one(context.Background(), &d.order[0], "q", count); err != nil {
				t.Errorf("count %d: %v", count, err)
			}
		})
	}
	wg.Wait()
}
