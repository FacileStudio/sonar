package server

import (
	"log/slog"
	"net/http"

	"github.com/FacileStudio/tronc/middleware"
	"github.com/FacileStudio/tronc/spa"
)

// securityConfig extends tronc's default policy with the hashes of the inline
// scripts adapter-static writes and Vision's tracker origin. Images come from
// the same-origin proxy and Vision's pixels from https, both of which tronc's
// default img-src allows, and the policy is enforced.
func securityConfig(clientDir string, log *slog.Logger) middleware.SecurityConfig {
	sources := []string{visionOrigin}
	if spa.Available(clientDir) {
		hashes, err := spa.InlineScriptHashes(clientDir)
		if err != nil {
			log.Error("failed to hash the client's inline scripts", slog.Any("error", err))
		}
		sources = append(sources, hashes...)
	}
	return middleware.SecurityConfig{CSPEnforce: true, CSPSources: map[string][]string{"script-src": sources}}
}

// visionOrigin serves the analytics tracker app.html loads on the production host.
const visionOrigin = "https://vision.facile.studio"

// noReferrer keeps the browser from sending a Referer from sonar's pages, so
// a result site never learns the query from the search page's URL.
func noReferrer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(middleware.HeaderReferrerPolicy, "no-referrer")
		next.ServeHTTP(w, r)
	})
}
