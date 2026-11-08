// Package keys lets a person bring their own engine API keys. Searches that
// run on them skip the daily quota. Keys are write-only: once saved, only the
// last four characters ever come back.
package keys

import (
	"net/http"

	"github.com/FacileStudio/sonar/internal/server/keyring"
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts the key routes behind sign-in. porte's session
// middleware already requires the CSRF header on cookie-authenticated writes.
func RegisterRoutes(router chi.Router, ring *keyring.Ring, requireAuth func(http.Handler) http.Handler) {
	h := &handler{ring: ring}
	router.Group(func(r chi.Router) {
		r.Use(requireAuth)
		r.Get("/keys", h.list)
		r.Put("/keys/{engine}", h.set)
		r.Delete("/keys/{engine}", h.remove)
	})
}
