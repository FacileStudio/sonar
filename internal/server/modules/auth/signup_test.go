package auth

import (
	"context"
	"database/sql"
	"slices"
	"testing"

	"github.com/FacileStudio/porte"
	portepg "github.com/FacileStudio/porte/pg"
	"github.com/FacileStudio/sonar/internal/server/servertest"
)

func TestNewAccountLeavesOnlyPinnedColumnsNull(t *testing.T) {
	db := servertest.DB(t)
	users := pictureUsers{UserStore: portepg.New(db).Users()}
	id, err := users.UpsertFromOIDC(context.Background(), porte.Claims{Provider: "https://sso.example", Subject: "sub-new",
		Email: "new@example.com", EmailVerified: true, Name: "New", Picture: "https://sso.example/avatars/sub-new"})
	if err != nil {
		t.Fatal(err)
	}
	null := nullColumns(t, db, id)
	if want := []string{"facile_id"}; !slices.Equal(null, want) {
		t.Fatalf("a new account leaves %v NULL, want %v", null, want)
	}
}

func nullColumns(t *testing.T, db *sql.DB, id int64) []string {
	t.Helper()
	rows, err := db.Query(`SELECT column_name FROM information_schema.columns
		WHERE table_schema = current_schema() AND table_name = 'porte_users' ORDER BY ordinal_position`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	var null []string
	for rows.Next() {
		var column string
		if err := rows.Scan(&column); err != nil {
			t.Fatal(err)
		}
		var isNull bool
		if err := db.QueryRow(`SELECT `+column+` IS NULL FROM porte_users WHERE id = $1`, id).Scan(&isNull); err != nil {
			t.Fatal(err)
		}
		if isNull {
			null = append(null, column)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return null
}
