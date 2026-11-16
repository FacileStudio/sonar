package admin

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/FacileStudio/porte"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

// Admins is who may open the admin routes. Emails is SONAR_ADMINS, the
// suite's allow-list of Registre sign-in addresses. A session is let in when
// its Registre roles hold Role, or when its account's verified email, which
// only a verified Registre claim can write, is on Emails. An empty list lets
// nobody in by email.
type Admins struct {
	DB     *sql.DB
	Emails []string
}

func (a Admins) allows(ctx context.Context, identity porte.Identity) (bool, error) {
	if identity.HasRole(Role) {
		return true, nil
	}
	if len(a.Emails) == 0 || identity.UserID == 0 {
		return false, nil
	}
	var email string
	err := a.DB.QueryRowContext(ctx, `SELECT email FROM porte_users WHERE id = $1 AND email_verified`,
		identity.UserID).Scan(&email)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return listed(a.Emails, email), nil
}

func listed(emails []string, email string) bool {
	want := strings.ToLower(strings.TrimSpace(email))
	if want == "" {
		return false
	}
	for _, admin := range emails {
		if strings.ToLower(strings.TrimSpace(admin)) == want {
			return true
		}
	}
	return false
}

func (a Admins) require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		identity, ok := porte.From(r.Context())
		if !ok {
			httpjson.WriteError(w, troncerrors.Unauthorized("missing auth"))
			return
		}
		allowed, err := a.allows(r.Context(), identity)
		if err != nil {
			httpjson.WriteError(w, troncerrors.Internal("failed to read the account", err))
			return
		}
		if !allowed {
			httpjson.WriteError(w, troncerrors.Forbidden("this page is for sonar's administrators"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
