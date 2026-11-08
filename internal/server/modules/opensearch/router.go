// Package opensearch lets a browser install sonar as a search engine: the
// OpenSearch description, and the /search entry point that answers an
// external !bang with a redirect before the app even loads.
package opensearch

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts /opensearch.xml and /search on the root router. app is
// the client the /search page falls through to.
func RegisterRoutes(router chi.Router, publicURL string, app http.Handler) {
	router.Get("/opensearch.xml", description(publicURL))
	router.Get("/search", entry(app))
}
