package search

import (
	"net/http"

	troncmiddleware "github.com/FacileStudio/tronc/middleware"
	"github.com/go-chi/chi/v5"
)

// Routes is what the module mounts: the service, the suggestion source and
// the sign-in middleware every route sits behind.
type Routes struct {
	Service     *Service
	Suggest     *Suggester
	RequireAuth func(http.Handler) http.Handler
	UserKey     func(*http.Request) string
}

// RegisterRoutes mounts the search API under router, every route behind
// sign-in and a per-user rate limit.
func RegisterRoutes(router chi.Router, routes Routes) {
	searches := troncmiddleware.Ratelimit(troncmiddleware.RatelimitConfig{Limit: 2, Burst: 20, KeyFunc: routes.UserKey})
	typing := troncmiddleware.Ratelimit(troncmiddleware.RatelimitConfig{Limit: 8, Burst: 40, KeyFunc: routes.UserKey})
	h := &handler{service: routes.Service, suggest: routes.Suggest}
	router.Group(func(r chi.Router) {
		r.Use(routes.RequireAuth)
		r.With(searches).Get("/search", h.search)
		r.With(searches).Post("/fetch", h.fetch)
		r.With(typing).Get("/suggest", h.suggestions)
		r.Get("/usage", h.usage)
	})
	router.Get("/bangs", h.bangs)
}
