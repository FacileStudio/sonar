// Package admin serves the operator's read-only views of the hosted service.
// Every route needs a session that Admins lets in: a Registre admin role, or
// a verified email on SONAR_ADMINS. Everyone else is refused.
package admin

import (
	"net/http"

	"github.com/FacileStudio/sonar/internal/server/enginestats"
	"github.com/go-chi/chi/v5"
)

// Role is the Registre role that opens the admin routes.
const Role = "admin"

// RegisterRoutes mounts the admin routes under router, behind sign-in and
// admins.
func RegisterRoutes(router chi.Router, stats *enginestats.Store, admins Admins,
	requireAuth func(http.Handler) http.Handler) {
	h := &handler{stats: stats}
	router.Route("/admin", func(r chi.Router) {
		r.Use(requireAuth, admins.require)
		r.Get("/engine-stats", h.engineStats)
	})
}
