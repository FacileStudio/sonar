package search

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/sonar/internal/server/enginestats"
)

type recorded struct {
	kind  string
	tried []enginestats.Attempt
}

type fakeStats struct {
	got  chan recorded
	fail error
}

func newFakeStats() *fakeStats { return &fakeStats{got: make(chan recorded, 8)} }

func (f *fakeStats) Record(_ context.Context, kind string, tried []enginestats.Attempt) error {
	f.got <- recorded{kind: kind, tried: tried}
	return f.fail
}

func (f *fakeStats) next(t *testing.T) recorded {
	t.Helper()
	select {
	case r := <-f.got:
		return r
	case <-time.After(2 * time.Second):
		t.Fatal("no engine outcome was recorded")
		return recorded{}
	}
}

func (f *fakeStats) none(t *testing.T) {
	t.Helper()
	select {
	case r := <-f.got:
		t.Fatalf("nothing should be recorded, got %+v", r)
	case <-time.After(50 * time.Millisecond):
	}
}

func attempts(pairs ...string) []enginestats.Attempt {
	var out []enginestats.Attempt
	for i := 0; i < len(pairs); i += 2 {
		out = append(out, enginestats.Attempt{Engine: pairs[i], Outcome: enginestats.Outcome(pairs[i+1])})
	}
	return out
}

type chainCase struct {
	name   string
	hits   map[string]int
	broken []string
	stuck  []string
	want   []enginestats.Attempt
}

func (tc chainCase) engines() *perEngine {
	engines := &perEngine{hits: tc.hits, broken: map[string]bool{}, stuck: map[string]bool{}}
	for _, name := range tc.broken {
		engines.broken[name] = true
	}
	for _, name := range tc.stuck {
		engines.stuck[name] = true
	}
	return engines
}

func chainCases() []chainCase {
	return []chainCase{
		{"answered", map[string]int{"[searxng]": 7}, nil, nil, attempts("searxng", "answered")},
		{"short then fallback", map[string]int{"[searxng]": 2, "[brave]": 8}, nil, nil,
			attempts("searxng", "short", "brave", "answered")},
		{"failed then fallback", map[string]int{"[brave]": 8}, []string{"[searxng]"}, nil,
			attempts("searxng", "failed", "brave", "answered")},
		{"timeout then fallback", map[string]int{"[brave]": 8}, nil, []string{"[searxng]"},
			attempts("searxng", "timed_out", "brave", "answered")},
		{"only a short answer", map[string]int{"[searxng]": 2}, []string{"[brave]", "[tavily]"}, nil,
			attempts("searxng", "answered", "brave", "failed", "tavily", "failed")},
		{"nothing", nil, []string{"[searxng]", "[brave]", "[tavily]"}, nil,
			attempts("searxng", "failed", "brave", "failed", "tavily", "failed")},
	}
}

func TestWebChainRecordsHowEachEngineAnswered(t *testing.T) {
	old := freeStageTimeout
	freeStageTimeout = 20 * time.Millisecond
	defer func() { freeStageTimeout = old }()
	for _, tc := range chainCases() {
		stats := newFakeStats()
		h := newHarness(10, nil)
		h.svc.Instance, h.svc.Stats = tc.engines(), stats
		h.svc.Web = Chain{Engines: []string{"searxng", "brave", "tavily"}, Enough: 5}
		response, _ := h.svc.Search(context.Background(), 1, Query{Q: "zanzibarquokka", Kind: KindWeb})
		got := stats.next(t)
		if got.kind != "web" || !reflect.DeepEqual(got.tried, tc.want) {
			t.Fatalf("%s: recorded %s %+v, want web %+v (%d results)", tc.name, got.kind, got.tried, tc.want, len(response.Web))
		}
	}
}

func TestImagesAndSummarySourcesAreRecordedUnderTheirKind(t *testing.T) {
	stats := newFakeStats()
	h := newHarness(10, nil)
	h.svc.Stats = stats
	if _, err := h.svc.Search(context.Background(), 1, Query{Q: "cats", Kind: KindImages}); err != nil {
		t.Fatal(err)
	}
	if got := stats.next(t); got.kind != "images" || !reflect.DeepEqual(got.tried, attempts("brave", "answered")) {
		t.Fatalf("images recorded %+v", got)
	}
	h.instance.fail = errors.New("brave is down")
	_, _ = h.svc.Search(context.Background(), 1, Query{Q: "cats", Kind: KindNews})
	if got := stats.next(t); got.kind != "news" || !reflect.DeepEqual(got.tried, attempts("brave", "failed")) {
		t.Fatalf("news recorded %+v", got)
	}
	h.svc.Instance = &perEngine{hits: map[string]int{"[tavily]": 6}}
	h.svc.Sources = Chain{Engines: []string{"tavily", "exa"}, Enough: 5}
	if _, err := h.svc.SummarySources(context.Background(), 1, "cats", 10); err != nil {
		t.Fatal(err)
	}
	if got := stats.next(t); got.kind != KindSummary || !reflect.DeepEqual(got.tried, attempts("tavily", "answered")) {
		t.Fatalf("summary recorded %+v", got)
	}
}

func TestSearchesOnAPersonsOwnKeysAreNotRecorded(t *testing.T) {
	stats := newFakeStats()
	h := newHarness(10, fakeKeys{"brave": "key"})
	h.svc.Stats = stats
	h.svc.Web = Chain{Engines: []string{"searxng", "brave"}, Enough: 5}
	for _, kind := range []Kind{KindWeb, KindImages} {
		if _, err := h.svc.Search(context.Background(), 1, Query{Q: "cats", Kind: kind}); err != nil {
			t.Fatal(err)
		}
	}
	stats.none(t)
}

func TestAFailingCounterNeverFailsASearchNorLogsTheQuery(t *testing.T) {
	var logs bytes.Buffer
	stats := newFakeStats()
	stats.fail = errors.New("the database is away")
	h := newHarness(10, nil)
	h.svc.Instance, h.svc.Stats = &perEngine{hits: map[string]int{"[searxng]": 7}}, stats
	h.svc.Logger = slog.New(slog.NewJSONHandler(&syncWriter{buf: &logs}, nil))
	h.svc.Web = Chain{Engines: []string{"searxng"}, Enough: 5}
	response, err := h.svc.Search(context.Background(), 1, Query{Q: "zanzibarquokka", Kind: KindWeb})
	if err != nil || len(response.Web) != 7 {
		t.Fatalf("%d results, err %v", len(response.Web), err)
	}
	stats.next(t)
	deadline := time.Now().Add(2 * time.Second)
	for !strings.Contains(read(&logs), "failed to count engine outcomes") {
		if time.Now().After(deadline) {
			t.Fatal("the counter failure was not logged")
		}
		time.Sleep(5 * time.Millisecond)
	}
	if strings.Contains(read(&logs), "zanzibarquokka") {
		t.Fatalf("the query reached the log: %s", read(&logs))
	}
}
