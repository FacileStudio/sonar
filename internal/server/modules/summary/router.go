// Package summary serves the paid AI summary of a result page. sonar holds no
// AI key: the summary is written by Bulle, the suite's AI hub, acting for the
// signed-in person, and only plans with ai.access may ask for one. The feature
// is off unless SONAR_AI_SUMMARY is true and Bulle is configured.
package summary

import (
	"net/http"

	troncmiddleware "github.com/FacileStudio/tronc/middleware"
	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts POST /summary behind sign-in and a per-user limit.
func RegisterRoutes(router chi.Router, service *Service, requireAuth func(http.Handler) http.Handler,
	userKey func(*http.Request) string) {
	limit := troncmiddleware.Ratelimit(troncmiddleware.RatelimitConfig{Limit: 0.5, Burst: 5, KeyFunc: userKey})
	router.With(requireAuth, limit).Post("/summary", service.handle)
}
