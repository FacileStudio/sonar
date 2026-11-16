// Package admin serves the operator's read-only views of the hosted service.
// Every route needs a session whose roles, as Registre grants them for sonar,
// hold the admin role; everyone else is refused.
package admin

import (
	"net/http"

	"github.com/FacileStudio/sonar/internal/server/enginestats"
	"github.com/go-chi/chi/v5"
)

// Role is the Registre role that opens the admin routes.
const Role = "admin"

// RegisterRoutes mounts the admin routes under router, behind sign-in and the
// admin role.
func RegisterRoutes(router chi.Router, stats *enginestats.Store, requireAuth func(http.Handler) http.Handler) {
	h := &handler{stats: stats}
	router.Route("/admin", func(r chi.Router) {
		r.Use(requireAuth, requireAdmin)
		r.Get("/engine-stats", h.engineStats)
	})
}
