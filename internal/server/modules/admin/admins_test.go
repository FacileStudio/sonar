package admin_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FacileStudio/porte"
	portepg "github.com/FacileStudio/porte/pg"
	"github.com/FacileStudio/porte/session"
	"github.com/FacileStudio/sonar/internal/server/enginestats"
	"github.com/FacileStudio/sonar/internal/server/modules/admin"
	"github.com/FacileStudio/sonar/internal/server/servertest"
	"github.com/go-chi/chi/v5"
)

const issuer = "https://sso.example"

type allowListCase struct {
	emails []string
	user   int64
	want   int
}

func verifiedUser(t *testing.T, db *sql.DB, email, subject string, verified bool) int64 {
	t.Helper()
	id := servertest.User(t, db, email, issuer, subject)
	if _, err := db.Exec(`UPDATE porte_users SET email_verified = $2 WHERE id = $1`, id, verified); err != nil {
		t.Fatal(err)
	}
	return id
}

func viaSession(t *testing.T, db *sql.DB, emails []string, userID int64) int {
	t.Helper()
	sessions, err := session.New(porte.Config{AppName: "Sonar"}, session.Deps{Sessions: portepg.New(db).Sessions()})
	if err != nil {
		t.Fatal(err)
	}
	router := chi.NewRouter()
	admin.RegisterRoutes(router, &enginestats.Store{DB: db}, admin.Admins{DB: db, Emails: emails}, sessions.RequireAuth)
	request := httptest.NewRequest(http.MethodGet, "/admin/engine-stats", nil)
	if userID != 0 {
		token, _, err := sessions.Issue(context.Background(), userID, "")
		if err != nil {
			t.Fatal(err)
		}
		request.AddCookie(&http.Cookie{Name: porte.SessionCookieName, Value: token})
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response.Code
}

func TestSonarAdminsOpensTheRouteToListedVerifiedEmailsOnly(t *testing.T) {
	db := servertest.DB(t)
	owner := verifiedUser(t, db, "owner@facile.test", "sub-owner", true)
	member := verifiedUser(t, db, "member@facile.test", "sub-member", true)
	squatter := verifiedUser(t, db, "squat@facile.test", "sub-squat", false)
	listed := []string{" Owner@Facile.test ", "squat@facile.test"}
	for name, c := range map[string]allowListCase{
		"listed":            {emails: listed, user: owner, want: http.StatusOK},
		"unlisted":          {emails: listed, user: member, want: http.StatusForbidden},
		"listed unverified": {emails: listed, user: squatter, want: http.StatusForbidden},
		"signed out":        {emails: listed, want: http.StatusUnauthorized},
		"empty list":        {user: owner, want: http.StatusForbidden},
		"empty list member": {user: member, want: http.StatusForbidden},
	} {
		if got := viaSession(t, db, c.emails, c.user); got != c.want {
			t.Errorf("%s: status %d, want %d", name, got, c.want)
		}
	}
}
