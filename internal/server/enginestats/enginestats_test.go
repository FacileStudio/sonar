package enginestats_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/FacileStudio/sonar/internal/server/enginestats"
	"github.com/FacileStudio/sonar/internal/server/servertest"
)

func record(t *testing.T, store *enginestats.Store, kind string, pairs ...string) {
	t.Helper()
	var tried []enginestats.Attempt
	for i := 0; i < len(pairs); i += 2 {
		tried = append(tried, enginestats.Attempt{Engine: pairs[i], Outcome: enginestats.Outcome(pairs[i+1])})
	}
	if err := store.Record(context.Background(), kind, tried); err != nil {
		t.Fatal(err)
	}
}

func TestRecordAddsUpPerDayKindAndEngine(t *testing.T) {
	now := time.Date(2026, 11, 16, 23, 30, 0, 0, time.UTC)
	store := &enginestats.Store{DB: servertest.DB(t), Now: func() time.Time { return now }}
	ctx := context.Background()
	record(t, store, "web", "searxng", "answered")
	record(t, store, "web", "searxng", "short", "brave", "answered")
	record(t, store, "web", "searxng", "timed_out", "ddg", "failed", "brave", "answered")
	record(t, store, "news", "brave", "failed")
	now = now.Add(time.Hour)
	record(t, store, "web", "searxng", "answered")
	got, err := store.Read(ctx, 2)
	want := []enginestats.Row{
		{Day: "2026-11-17", Kind: "web", Engine: "searxng", Tried: 1, Answered: 1},
		{Day: "2026-11-16", Kind: "news", Engine: "brave", Tried: 1, Failed: 1},
		{Day: "2026-11-16", Kind: "web", Engine: "brave", Tried: 2, Answered: 2},
		{Day: "2026-11-16", Kind: "web", Engine: "ddg", Tried: 1, Failed: 1},
		{Day: "2026-11-16", Kind: "web", Engine: "searxng", Tried: 3, Answered: 1, Short: 1, TimedOut: 1},
	}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("rows %+v, want %+v, err %v", got, want, err)
	}
	if today, _ := store.Read(ctx, 1); len(today) != 1 {
		t.Fatalf("one day back must hold today only, got %+v", today)
	}
}

func TestTheTableHoldsCountersOnly(t *testing.T) {
	var got string
	err := servertest.DB(t).QueryRow(`SELECT string_agg(column_name, ' ' ORDER BY column_name)
		FROM information_schema.columns
		WHERE table_schema = current_schema() AND table_name = 'engine_stats'`).Scan(&got)
	if want := "answered day engine failed kind short timed_out tried"; err != nil || got != want {
		t.Fatalf("engine_stats columns %q, want %q, err %v: a column must not identify a search or a person", got, want, err)
	}
}
