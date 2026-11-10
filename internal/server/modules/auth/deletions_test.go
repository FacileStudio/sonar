package auth

import (
	"context"
	"database/sql"
	"testing"

	"github.com/FacileStudio/porte/directory"
	portepg "github.com/FacileStudio/porte/pg"
	"github.com/FacileStudio/sonar/internal/server/keyring"
	"github.com/FacileStudio/sonar/internal/server/servertest"
)

var personTables = []string{"porte_identities", "porte_sessions", "porte_login_codes", "porte_magic_links",
	"search_usage", "engine_keys"}

func TestErasureDeletesEveryRowOfThePerson(t *testing.T) {
	for _, stored := range []string{"https://sso.example", "https://sso.example/"} {
		t.Run(stored, func(t *testing.T) {
			db := servertest.DB(t)
			ctx := context.Background()
			gone := personWithData(t, db, "gone@example.com", stored, "sub-gone")
			kept := personWithData(t, db, "kept@example.com", stored, "sub-kept")
			deletions := newDeletions(db, portepg.New(db), directory.Client{Issuer: "https://sso.example"}, nil)
			if err := deletions.OnPersonDeleted(ctx, deletions.Client.Issuer, "sub-gone"); err != nil {
				t.Fatal(err)
			}
			if err := deletions.Erase(ctx, deletions.Client.Issuer, "sub-gone"); err != nil {
				t.Fatal(err)
			}
			if n := rowsOf(t, db, gone); n != 0 {
				t.Fatalf("%d rows still reference the erased person", n)
			}
			if n := rowsOf(t, db, kept); n != 1+len(personTables) {
				t.Fatalf("the other person kept %d rows, want %d", n, 1+len(personTables))
			}
		})
	}
}

func TestPersonTablesAreEveryTableReferencingUsers(t *testing.T) {
	db := servertest.DB(t)
	rows, err := db.Query(`SELECT DISTINCT tc.table_name FROM information_schema.table_constraints tc
		JOIN information_schema.constraint_column_usage cu ON cu.constraint_name = tc.constraint_name
		 AND cu.constraint_schema = tc.constraint_schema
		WHERE tc.constraint_type = 'FOREIGN KEY' AND cu.table_name = 'porte_users'
		  AND tc.table_schema = current_schema() ORDER BY 1`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	found := map[string]bool{}
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			t.Fatal(err)
		}
		found[table] = true
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(found) != len(personTables) {
		t.Fatalf("tables referencing porte_users: %v, erasure test covers %v", found, personTables)
	}
	for _, table := range personTables {
		if !found[table] {
			t.Fatalf("%s no longer references porte_users", table)
		}
	}
}

func personWithData(t *testing.T, db *sql.DB, email, provider, subject string) int64 {
	t.Helper()
	id := servertest.User(t, db, email, provider, subject)
	sealer, err := portepg.NewTokenSealer([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	if err := (&keyring.Ring{DB: db, Sealer: sealer}).Set(context.Background(), id, "brave", "test-key-not-real"); err != nil {
		t.Fatal(err)
	}
	for _, statement := range []string{
		`INSERT INTO porte_sessions (token_hash, user_id) VALUES ('session-' || $2, $1)`,
		`INSERT INTO porte_login_codes (code_hash, user_id, expires_at) VALUES ('code-' || $2, $1, now())`,
		`INSERT INTO porte_magic_links (token_hash, user_id, created_at, expires_at) VALUES ('link-' || $2, $1, now(), now())`,
		`INSERT INTO search_usage (user_id, day, count) SELECT $1, current_date, 3 WHERE $2 <> ''`,
	} {
		if _, err := db.Exec(statement, id, subject); err != nil {
			t.Fatal(err)
		}
	}
	return id
}

func rowsOf(t *testing.T, db *sql.DB, userID int64) int {
	t.Helper()
	var total int
	if err := db.QueryRow(`SELECT count(*) FROM porte_users WHERE id = $1`, userID).Scan(&total); err != nil {
		t.Fatal(err)
	}
	for _, table := range personTables {
		var n int
		if err := db.QueryRow(`SELECT count(*) FROM `+table+` WHERE user_id = $1`, userID).Scan(&n); err != nil {
			t.Fatal(err)
		}
		total += n
	}
	return total
}
