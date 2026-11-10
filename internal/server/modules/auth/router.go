package auth

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/directory"
	"github.com/FacileStudio/porte/kit"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

// Features tells the client which optional parts this deployment runs.
type Features struct {
	SignIn  bool `json:"sign_in"`
	Summary bool `json:"summary"`
	Keys    bool `json:"keys"`
	Suggest bool `json:"suggest"`
}

type sessionUser struct {
	Email     string `json:"email"`
	Name      string `json:"name,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
}

type sessionResponse struct {
	User     *sessionUser `json:"user"`
	Features Features     `json:"features"`
}

// RegisterRoutes mounts porte's routes, GET /auth/session and, with an
// issuer, porte/kit's GET /me/invitations under the API router.
func RegisterRoutes(router chi.Router, kits *Kits, features Features) {
	kits.Sessions.Mount(router)
	kits.OIDC.Mount(router)
	router.With(kits.Sessions.Optional).Get("/auth/session", handleSession(kits.DB, features, kits.Issuer))
	if kits.Invitations != nil {
		router.With(kits.Sessions.RequireAuth).Method(http.MethodGet, kit.InvitationsPath, kits.Invitations)
	}
}

// MountRegistre mounts Registre's directory nudge and deletion notices on the
// root router, at the fixed paths porte/directory defines.
func MountRegistre(root chi.Router, kits *Kits) {
	if kits.Syncer != nil {
		root.Method(http.MethodPost, directory.SyncPath, kits.Syncer.Handler())
	}
	if kits.Deletions != nil {
		root.Method(http.MethodPost, directory.DeletionsPath, kits.Deletions.Handler())
	}
}

func handleSession(db *sql.DB, features Features, issuer string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		identity, ok := porte.From(r.Context())
		if !ok {
			httpjson.WriteJSON(w, http.StatusOK, sessionResponse{Features: features})
			return
		}
		user := &sessionUser{Email: identity.Email, Name: identity.Name}
		_ = db.QueryRowContext(r.Context(), `SELECT email, name, avatar_url FROM porte_users WHERE id = $1`,
			identity.UserID).Scan(&user.Email, &user.Name, &user.AvatarURL)
		if user.AvatarURL == "" && strings.HasPrefix(issuer, "https://") {
			var subject string
			if err := db.QueryRowContext(r.Context(),
				`SELECT subject FROM porte_identities WHERE user_id = $1 AND rtrim(provider, '/') = $2 LIMIT 1`,
				identity.UserID, strings.TrimRight(issuer, "/")).Scan(&subject); err == nil && subject != "" {
				user.AvatarURL = strings.TrimRight(issuer, "/") + "/avatars/" + subject
			}
		}
		httpjson.WriteJSON(w, http.StatusOK, sessionResponse{User: user, Features: features})
	}
}
