package plans

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	"github.com/FacileStudio/porte/directory"
	"github.com/FacileStudio/sonar/internal/server/servertest"
	troncerrors "github.com/FacileStudio/tronc/errors"
)

const issuer = "https://sso.example"

func snapshot(revision int64, people map[string]string) directory.Snapshot {
	snap := directory.Snapshot{Version: revision, EntitlementsRevision: revision}
	for sub, values := range people {
		snap.People = append(snap.People, directory.Person{Subject: sub, Entitlements: decode(values)})
	}
	return snap
}

func decode(values string) map[string]json.RawMessage {
	out := map[string]json.RawMessage{}
	_ = json.Unmarshal([]byte(values), &out)
	return out
}

func TestPlansFollowEntitlementsAndSurviveARestart(t *testing.T) {
	db := servertest.DB(t)
	free := servertest.User(t, db, "free@example.com", issuer, "sub-free")
	pro := servertest.User(t, db, "pro@example.com", issuer, "sub-pro")
	gate := New(db, issuer, slog.Default())
	ctx := context.Background()
	if limit, _ := gate.Daily(ctx, pro); limit != DefaultDaily {
		t.Fatalf("before any snapshot the quota must be the free default, got %d", limit)
	}
	if err := gate.CheckAI(ctx, pro); troncerrors.Status(err) != 402 {
		t.Fatalf("before any snapshot AI must be refused with 402, got %v", err)
	}
	err := gate.Apply(ctx, snapshot(3, map[string]string{"sub-pro": `{"sonar.searches_daily":500,"ai.access":true}`}))
	if err != nil {
		t.Fatal(err)
	}
	restarted := New(db, issuer, slog.Default())
	if err := restarted.Load(ctx); err != nil {
		t.Fatal(err)
	}
	checkPlans(t, gate, free, pro)
	checkPlans(t, restarted, free, pro)
}

func checkPlans(t *testing.T, g *Gate, free, pro int64) {
	t.Helper()
	ctx := context.Background()
	if limit, _ := g.Daily(ctx, pro); limit != 500 {
		t.Fatalf("pro quota = %d", limit)
	}
	if limit, _ := g.Daily(ctx, free); limit != DefaultDaily {
		t.Fatalf("free quota = %d", limit)
	}
	if g.CheckAI(ctx, pro) != nil || !errors.Is(g.CheckAI(ctx, free), ErrAIAccess) {
		t.Fatal("ai.access must follow the plan")
	}
}

func TestUnlimitedWinsAcrossIdentities(t *testing.T) {
	db := servertest.DB(t)
	user := servertest.User(t, db, "two@example.com", issuer, "sub-one")
	if _, err := db.Exec(`INSERT INTO porte_identities (user_id, provider, subject) VALUES ($1, $2, 'sub-two')`,
		user, issuer); err != nil {
		t.Fatal(err)
	}
	gate := New(db, issuer, slog.Default())
	err := gate.Apply(context.Background(), snapshot(1, map[string]string{
		"sub-one": `{"sonar.searches_daily":500}`, "sub-two": `{"sonar.searches_daily":-1}`,
	}))
	if err != nil {
		t.Fatal(err)
	}
	if limit, _ := gate.Daily(context.Background(), user); limit != -1 {
		t.Fatalf("limit = %d, want unlimited", limit)
	}
}
