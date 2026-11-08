// Package servertest gives the server's tests a migrated PostgreSQL schema of
// their own. It reads TEST_DATABASE_URL and skips, loudly, without it.
package servertest

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"net/url"
	"os"
	"testing"

	"github.com/FacileStudio/sonar/internal/server/migrations"
	"github.com/FacileStudio/tronc/migrate"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// DB opens a fresh schema, migrates it and drops it when the test ends.
func DB(t *testing.T) *sql.DB {
	t.Helper()
	base := os.Getenv("TEST_DATABASE_URL")
	if base == "" {
		t.Skip("TEST_DATABASE_URL is unset: database tests skipped")
	}
	suffix := make([]byte, 6)
	_, _ = rand.Read(suffix)
	schema := "sonar_test_" + hex.EncodeToString(suffix)
	admin := open(t, base)
	if _, err := admin.Exec(`CREATE SCHEMA ` + schema); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = admin.Exec(`DROP SCHEMA ` + schema + ` CASCADE`)
		_ = admin.Close()
	})
	db := open(t, withSearchPath(t, base, schema))
	if err := migrate.Run(context.Background(), migrate.Config{DB: db, FS: migrations.FS}); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	return db
}

func open(t *testing.T, dsn string) *sql.DB {
	t.Helper()
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func withSearchPath(t *testing.T, base, schema string) string {
	t.Helper()
	parsed, err := url.Parse(base)
	if err != nil {
		t.Fatal(err)
	}
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

// User inserts a porte user with one Registre identity and returns its id.
func User(t *testing.T, db *sql.DB, email, issuer, subject string) int64 {
	t.Helper()
	var id int64
	if err := db.QueryRow(`INSERT INTO porte_users (email) VALUES ($1) RETURNING id`, email).Scan(&id); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO porte_identities (user_id, provider, subject) VALUES ($1, $2, $3)`,
		id, issuer, subject); err != nil {
		t.Fatal(err)
	}
	return id
}
