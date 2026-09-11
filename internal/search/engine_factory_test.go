// Package search's engine factory: nothing here touches the network, so the
// config-to-engine expansion is safe to pin in a plain unit test.
package search

import (
	"testing"
	"time"

	"github.com/FacileStudio/sonar/internal/config"
	"github.com/FacileStudio/sonar/internal/httpc"
)

func TestDefaultHasNoSearxngAndNoFuret(t *testing.T) {
	t.Helper()
	cfg := config.Default()
	if len(cfg.Searxng) != 0 {
		t.Fatalf("default searxng = %+v, want an empty list", cfg.Searxng)
	}
	if cfg.Engines["furet"] != nil {
		t.Fatal("default should not carry a furet engine")
	}
}

// A single configured instance is named plainly so it reads cleanly in results
// and cache keys; several get index suffixes so their caches and breakers stay
// distinct instead of the second serving the first's cached results.
func TestSearxngListBecomesDistinctEngineUnits(t *testing.T) {
	t.Helper()
	cfg := config.Default()
	cfg.Searxng = []string{"https://a.example/search", "https://b.example"}
	built := buildEngineUnits(cfg, httpc.Client(5*time.Second))
	if len(built) != 2 {
		t.Fatalf("built %d units, want 2", len(built))
	}
	if built[0].eng.Name() != "searxng1" || built[1].eng.Name() != "searxng2" {
		t.Fatalf("names = %q, %q; want searxng1, searxng2", built[0].eng.Name(), built[1].eng.Name())
	}
	if built[0].prio != built[1].prio || built[0].prio != cfg.SearxngPriority {
		t.Fatalf("prio = %d, want searxngPriority %d", built[0].prio, cfg.SearxngPriority)
	}
}
