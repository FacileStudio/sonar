package keyring

import (
	"context"
	"errors"
	"strings"
	"testing"

	portepg "github.com/FacileStudio/porte/pg"
	"github.com/FacileStudio/sonar/internal/server/servertest"
)

const secret = "BSA-very-secret-key-1234"

func savedRing(t *testing.T) (*Ring, int64) {
	t.Helper()
	db := servertest.DB(t)
	user := servertest.User(t, db, "k@example.com", "https://sso.example", "sub-k")
	sealer, err := portepg.NewTokenSealer([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	ring := &Ring{DB: db, Sealer: sealer}
	if err := ring.Set(context.Background(), user, "brave", secret); err != nil {
		t.Fatal(err)
	}
	return ring, user
}

func TestKeysAreSealedAtRestAndOnlyHintsComeBack(t *testing.T) {
	ring, user := savedRing(t)
	db, ctx := ring.DB, context.Background()
	var stored string
	if err := db.QueryRow(`SELECT sealed FROM engine_keys WHERE user_id = $1`, user).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(stored, "pv1:") || strings.Contains(stored, secret) {
		t.Fatalf("key stored unsealed: %q", stored)
	}
	saved, _ := ring.List(ctx, user)
	if len(saved) != 1 || saved[0].Hint != "1234" {
		t.Fatalf("list = %+v", saved)
	}
	keys, _ := ring.Keys(ctx, user)
	if keys["brave"] != secret {
		t.Fatal("the sealed key did not open")
	}
	if err := ring.Delete(ctx, user, "brave"); err != nil {
		t.Fatal(err)
	}
	if keys, _ := ring.Keys(ctx, user); len(keys) != 0 {
		t.Fatal("deleted key still opens")
	}
}

func TestValidateRefusesUnknownEnginesAndJunk(t *testing.T) {
	for _, c := range []struct{ engine, key string }{
		{"google", "abcdefghijk"}, {"brave", "short"}, {"brave", "has space in it"}, {"brave", "tab\tkey123"},
	} {
		if Validate(c.engine, c.key) == nil {
			t.Errorf("Validate(%q, %q) passed", c.engine, c.key)
		}
	}
	if err := (&Ring{}).Set(context.Background(), 1, "brave", "abcdefghijk"); !errors.Is(err, ErrDisabled) {
		t.Fatalf("a ring without a sealer must refuse, got %v", err)
	}
}
