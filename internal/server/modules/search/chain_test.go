package search

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/FacileStudio/sonar/websearch"
)

type perEngine struct {
	fakeEngines
	hits   map[string]int
	broken map[string]bool
	stuck  map[string]bool
	called []string
}

func (p *perEngine) Search(ctx context.Context, _ string, _ int, include []string) ([]websearch.Result, error) {
	name := fmt.Sprint(include)
	p.called = append(p.called, name)
	if p.stuck[name] {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	if p.broken[name] {
		return nil, errors.New("down")
	}
	var out []websearch.Result
	for i := range p.hits[name] {
		out = append(out, websearch.Result{URL: fmt.Sprintf("https://%s.example/%d", name, i), Engine: name})
	}
	return out, nil
}

func TestWebChainCallsTavilyOnlyWhenBraveFallsShort(t *testing.T) {
	cases := []struct {
		name   string
		brave  int
		broken bool
		want   []string
	}{
		{"enough", 8, false, []string{"[brave]"}},
		{"few", 2, false, []string{"[brave]", "[tavily]"}},
		{"down", 0, true, []string{"[brave]", "[tavily]"}},
	}
	for _, tc := range cases {
		engines := &perEngine{hits: map[string]int{"[brave]": tc.brave, "[tavily]": 4}, broken: map[string]bool{"[brave]": tc.broken}}
		h := newHarness(10, nil)
		h.svc.Instance = engines
		h.svc.Web = Chain{Engines: []string{"brave", "tavily"}, Enough: 5}
		response, err := h.svc.Search(context.Background(), 1, Query{Q: "cats", Kind: KindWeb})
		if err != nil || !reflect.DeepEqual(engines.called, tc.want) || len(response.Web) == 0 {
			t.Fatalf("%s: called %v, want %v, %d results, err %v", tc.name, engines.called, tc.want, len(response.Web), err)
		}
	}
}

func TestSummarySourcesNeverReachBrave(t *testing.T) {
	engines := &perEngine{hits: map[string]int{"[brave]": 9, "[tavily]": 1, "[exa]": 6}}
	h := newHarness(10, nil)
	h.svc.Instance = engines
	h.svc.Sources = Chain{Engines: []string{"brave", "tavily", "exa", "parallel"}, Enough: 5}.Without("brave")
	hits, err := h.svc.SummarySources(context.Background(), 1, "cats", 10)
	if err != nil || len(hits) != 7 || !reflect.DeepEqual(engines.called, []string{"[tavily]", "[exa]"}) || h.counter.used != 1 {
		t.Fatalf("called %v, %d hits, used %d, err %v", engines.called, len(hits), h.counter.used, err)
	}
}

func TestWebChainTriesFreeEnginesBeforePaidOnes(t *testing.T) {
	cases := []struct {
		name string
		free int
		down bool
		want []string
	}{
		{"free is enough", 7, false, []string{"[searxpool]"}},
		{"free is thin", 2, false, []string{"[searxpool]", "[brave]"}},
		{"free is down", 0, true, []string{"[searxpool]", "[brave]"}},
	}
	for _, tc := range cases {
		engines := &perEngine{hits: map[string]int{"[searxpool]": tc.free, "[brave]": 8, "[tavily]": 4},
			broken: map[string]bool{"[searxpool]": tc.down}}
		h := newHarness(10, nil)
		h.svc.Instance = engines
		h.svc.Web = Chain{Engines: []string{"searxpool", "brave", "tavily"}, Enough: 5}
		if _, err := h.svc.Search(context.Background(), 1, Query{Q: "cats", Kind: KindWeb}); err != nil || !reflect.DeepEqual(engines.called, tc.want) {
			t.Fatalf("%s: called %v, want %v, err %v", tc.name, engines.called, tc.want, err)
		}
	}
}

func TestWebChainCutsASlowFreeEngineShort(t *testing.T) {
	old := freeStageTimeout
	freeStageTimeout = 20 * time.Millisecond
	defer func() { freeStageTimeout = old }()
	engines := &perEngine{hits: map[string]int{"[brave]": 8},
		stuck: map[string]bool{"[searxng]": true, "[ddg]": true}}
	h := newHarness(10, nil)
	h.svc.Instance = engines
	h.svc.Web = Chain{Engines: []string{"searxng", "ddg", "brave", "tavily"}, Enough: 5}
	start := time.Now()
	response, err := h.svc.Search(context.Background(), 1, Query{Q: "cats", Kind: KindWeb})
	want := []string{"[searxng]", "[ddg]", "[brave]"}
	if err != nil || !reflect.DeepEqual(engines.called, want) || len(response.Web) != 8 {
		t.Fatalf("called %v, want %v, %d results, err %v", engines.called, want, len(response.Web), err)
	}
	if took := time.Since(start); took > 2*time.Second {
		t.Fatalf("two stuck free engines held the search for %s", took)
	}
}

func TestWebChainLimitsAFreeEngineEvenWhenItIsLast(t *testing.T) {
	old := freeStageTimeout
	freeStageTimeout = 20 * time.Millisecond
	defer func() { freeStageTimeout = old }()
	engines := &perEngine{stuck: map[string]bool{"[ddg]": true}}
	h := newHarness(10, nil)
	h.svc.Instance = engines
	h.svc.Web = Chain{Engines: []string{"ddg"}, Enough: 5}
	start := time.Now()
	if _, err := h.svc.Search(context.Background(), 1, Query{Q: "cats", Kind: KindWeb}); err == nil {
		t.Fatal("a stuck engine with nothing behind it must fail the search")
	}
	if took := time.Since(start); took > 2*time.Second {
		t.Fatalf("a stuck free engine held the search for %s", took)
	}
}
