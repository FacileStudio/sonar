package enginestats_test

import (
	"testing"

	"github.com/FacileStudio/sonar/internal/server/enginestats"
)

func TestSummarizeSplitsAnsweredSearchesBetweenFreeAndPaid(t *testing.T) {
	rows := []enginestats.Row{
		{Day: "2026-11-16", Kind: "web", Engine: "searxng", Free: true, Tried: 10, Answered: 6, Short: 2, Failed: 1, TimedOut: 1},
		{Day: "2026-11-15", Kind: "web", Engine: "searxng", Free: true, Tried: 2, Answered: 1, Failed: 1},
		{Day: "2026-11-16", Kind: "web", Engine: "ddg", Free: true, Tried: 4, Failed: 4},
		{Day: "2026-11-16", Kind: "web", Engine: "brave", Tried: 5, Answered: 3, Failed: 2},
		{Day: "2026-11-16", Kind: "news", Engine: "brave", Tried: 2, Failed: 2},
	}
	got := enginestats.Summarize(rows)
	if len(got) != 2 || got[0].Kind != "news" || got[1].Kind != "web" {
		t.Fatalf("kinds %+v", got)
	}
	if news := got[0]; news.Answered != 0 || news.FreeShare != 0 || news.PaidShare != 0 {
		t.Fatalf("news with no answer must have zero shares, got %+v", news)
	}
	web := got[1]
	if web.Answered != 10 || web.FreeAnswered != 7 || web.PaidAnswered != 3 || web.FreeShare != 0.7 || web.PaidShare != 0.3 {
		t.Fatalf("web totals %+v", web)
	}
	if len(web.Engines) != 3 || web.Engines[0].Engine != "brave" || web.Engines[0].Share != 0.3 {
		t.Fatalf("web engines %+v", web.Engines)
	}
	if searxng := web.Engines[2]; searxng.Tried != 12 || searxng.Answered != 7 || searxng.Share != 0.7 || !searxng.Free {
		t.Fatalf("searxng total %+v", searxng)
	}
}
