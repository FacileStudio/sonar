package server

import (
	"database/sql"
	"net/http"

	"github.com/FacileStudio/sonar/internal/server/docs"
	"github.com/FacileStudio/sonar/internal/server/modules/auth"
	"github.com/FacileStudio/sonar/internal/server/modules/img"
	"github.com/FacileStudio/sonar/internal/server/modules/keys"
	"github.com/FacileStudio/sonar/internal/server/modules/mcp"
	"github.com/FacileStudio/sonar/internal/server/modules/opensearch"
	"github.com/FacileStudio/sonar/internal/server/modules/search"
	"github.com/FacileStudio/sonar/internal/server/modules/summary"
	"github.com/FacileStudio/tronc/apiref"
	"github.com/FacileStudio/tronc/health"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/FacileStudio/tronc/httpx"
	"github.com/FacileStudio/tronc/spa"
	"github.com/go-chi/chi/v5"
)

// buildRouter lays out the tree: /health, /version, /docs, /api, /mcp and
// its metadata when the MCP server is on, /img, /opensearch.xml, /search and
// the client.
func buildRouter(a *app, cfg Config, opts Options, db *sql.DB) chi.Router {
	clientDir := spa.DirFromEnv()
	router := httpx.NewRouter(httpx.Config{
		Security: securityConfig(clientDir, a.log), TrustedProxies: cfg.TrustedProxies,
		CDNProxies: cfg.CDNProxies, CDNHeader: cfg.CDNHeader, Logger: a.log,
	})
	health.Mount(router, health.DB(db))
	router.Get("/version", versionHandler(opts.Version))
	apiref.Mount(router, docs.Reference(opts.Version))
	auth.MountRegistre(router, a.kits)
	requireAuth := a.kits.Sessions.RequireAuth
	router.Route("/api", func(r chi.Router) {
		r.Get("/version", versionHandler(opts.Version))
		auth.RegisterRoutes(r, a.kits, a.features(cfg))
		search.RegisterRoutes(r, search.Routes{Service: a.search, Suggest: a.suggest, RequireAuth: requireAuth,
			UserKey: userKey})
		keys.RegisterRoutes(r, a.ring, requireAuth)
		summary.RegisterRoutes(r, a.summary, requireAuth, userKey)
	})
	if a.mcp != nil {
		mcp.RegisterRoutes(router, a.mcp)
	}
	img.RegisterRoutes(router, img.NewProxy(a.signer), requireAuth)
	client := clientHandler(clientDir)
	opensearch.RegisterRoutes(router, cfg.PublicURL, client)
	router.Handle("/*", client)
	return router
}

func (a *app) features(cfg Config) auth.Features {
	return auth.Features{
		SignIn:  cfg.OIDC.Issuer != "",
		Summary: a.summary.Writer != nil,
		Keys:    a.ring.Enabled(),
		Suggest: a.suggest != nil,
	}
}

func clientHandler(dir string) http.Handler {
	if !spa.Available(dir) {
		return http.NotFoundHandler()
	}
	return noReferrer(spa.Handler(spa.Config{Dir: dir}))
}

func versionHandler(version string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		httpjson.WriteJSON(w, http.StatusOK, map[string]string{"version": version})
	}
}
