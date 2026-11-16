package admin_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/sonar/internal/server/enginestats"
	"github.com/FacileStudio/sonar/internal/server/modules/admin"
	"github.com/FacileStudio/sonar/internal/server/modules/search"
	"github.com/FacileStudio/sonar/internal/server/servertest"
	"github.com/FacileStudio/sonar/websearch"
	"github.com/go-chi/chi/v5"
)

const secret = "zanzibarquokka"

type stubEngines struct{ search.Engines }

func (stubEngines) Search(_ context.Context, query string, _ int, include []string) ([]websearch.Result, error) {
	n := 8
	if include[0] == "searxng" && strings.Contains(query, "thin") {
		n = 2
	}
	out := make([]websearch.Result, n)
	for i := range out {
		out[i] = websearch.Result{URL: fmt.Sprintf("https://%s.example/%d", include[0], i), Engine: include[0]}
	}
	return out, nil
}

type stubPlans struct{ search.Plans }

func (stubPlans) Daily(context.Context, int64) (int64, error) { return -1, nil }

type stubQuota struct{ search.Quota }

func (stubQuota) Take(context.Context, int64, int64) (bool, error) { return true, nil }

type stubKeys struct{}

func (stubKeys) Keys(context.Context, int64) (map[string]string, error) { return nil, nil }

func signedInAs(identity porte.Identity) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r.WithContext(porte.WithIdentity(r.Context(), identity)))
		})
	}
}

func get(stats *enginestats.Store, identity porte.Identity, target string) *httptest.ResponseRecorder {
	router := chi.NewRouter()
	admin.RegisterRoutes(router, stats, admin.Admins{DB: stats.DB}, signedInAs(identity))
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, target, nil))
	return response
}

func searchThrice(t *testing.T, db *sql.DB, stats *enginestats.Store) {
	t.Helper()
	service := &search.Service{Instance: stubEngines{}, Plans: stubPlans{}, Quota: stubQuota{}, Keys: stubKeys{},
		Stats: stats, Logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		Web: search.Chain{Engines: []string{"searxng", "brave"}, Enough: 5}}
	for _, query := range []string{secret + " one", secret + " two", secret + " thin"} {
		if _, err := service.Search(context.Background(), 1, search.Query{Q: query, Kind: search.KindWeb}); err != nil {
			t.Fatal(err)
		}
	}
	deadline := time.Now().Add(3 * time.Second)
	for tried := 0; tried < 4; {
		if time.Now().After(deadline) {
			t.Fatalf("the counters hold %d engine turns, want 4", tried)
		}
		time.Sleep(10 * time.Millisecond)
		_ = db.QueryRow(`SELECT coalesce(sum(tried), 0) FROM engine_stats`).Scan(&tried)
	}
}

func TestEngineStatsCountsSearchesAndNeverHoldsTheQuery(t *testing.T) {
	db := servertest.DB(t)
	stats := &enginestats.Store{DB: db}
	searchThrice(t, db, stats)
	response := get(stats, porte.Identity{UserID: 1, Roles: []string{admin.Role}}, "/admin/engine-stats?days=7")
	if response.Code != http.StatusOK {
		t.Fatalf("status %d: %s", response.Code, response.Body)
	}
	var stored string
	if err := db.QueryRow(`SELECT string_agg(engine_stats::text, ' ') FROM engine_stats`).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	for name, text := range map[string]string{"the table": stored, "the route": response.Body.String()} {
		if strings.Contains(text, secret) || strings.Contains(text, "example") {
			t.Fatalf("%s holds a query or a result URL: %s", name, text)
		}
	}
	assertWebPeriod(t, response.Body.Bytes())
}

func assertWebPeriod(t *testing.T, body []byte) {
	t.Helper()
	var got admin.EngineStats
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if got.Days != 7 || len(got.Rows) != 2 || len(got.Period) != 1 {
		t.Fatalf("report %+v", got)
	}
	web := got.Period[0]
	if web.Kind != "web" || web.Answered != 3 || web.FreeAnswered != 2 || web.PaidAnswered != 1 {
		t.Fatalf("web period %+v", web)
	}
	brave, searxng := web.Engines[0], web.Engines[1]
	if brave.Free || brave.Tried != 1 || brave.Answered != 1 || !searxng.Free || searxng.Tried != 3 ||
		searxng.Answered != 2 || searxng.Short != 1 {
		t.Fatalf("engines %+v", web.Engines)
	}
}

func TestEngineStatsRefusesEveryoneButAdmins(t *testing.T) {
	stats := &enginestats.Store{DB: servertest.DB(t)}
	for name, identity := range map[string]porte.Identity{
		"no role":      {UserID: 2},
		"another role": {UserID: 3, Roles: []string{"member", "administrator"}},
	} {
		response := get(stats, identity, "/admin/engine-stats")
		if response.Code != http.StatusForbidden || strings.Contains(response.Body.String(), "rows") {
			t.Fatalf("%s: status %d, body %s", name, response.Code, response.Body)
		}
	}
	admin := porte.Identity{UserID: 1, Roles: []string{admin.Role}}
	if response := get(stats, admin, "/admin/engine-stats?days=0"); response.Code != http.StatusBadRequest {
		t.Fatalf("days=0: status %d", response.Code)
	}
	if response := get(stats, admin, "/admin/engine-stats"); response.Code != http.StatusOK {
		t.Fatalf("an admin was refused: %d", response.Code)
	}
}
