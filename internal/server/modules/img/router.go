// Package img is sonar's image proxy. Result thumbnails load through it, so a
// person's browser never contacts the image host and the page's CSP keeps
// img-src to its own origin. It serves only URLs sonar signed, only images,
// only from the public internet, and only up to a size cap.
package img

import (
	"net/http"
	"time"

	"github.com/FacileStudio/sonar/internal/netguard"
	"github.com/FacileStudio/sonar/internal/server/imgsig"
	"github.com/go-chi/chi/v5"
)

const (
	maxBytes = 5 << 20
	timeout  = 10 * time.Second
)

// Proxy fetches signed image URLs through the public-only guard.
type Proxy struct {
	Signer *imgsig.Signer
	Client *http.Client
}

// NewProxy builds a Proxy whose client dials only checked public addresses
// and keeps no cookies, so no image host can follow anyone across requests.
func NewProxy(signer *imgsig.Signer) *Proxy {
	client := (&netguard.Guard{}).Client(timeout)
	client.Jar = nil
	return &Proxy{Signer: signer, Client: client}
}

// RegisterRoutes mounts GET /img behind sign-in.
func RegisterRoutes(router chi.Router, proxy *Proxy, requireAuth func(http.Handler) http.Handler) {
	router.With(requireAuth).Get("/img", proxy.serve)
}
