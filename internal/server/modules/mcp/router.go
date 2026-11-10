package mcp

import "github.com/go-chi/chi/v5"

// RegisterRoutes mounts POST /mcp and its RFC 9728 metadata on router, the
// root router outside /api.
func RegisterRoutes(router chi.Router, service *Service) {
	service.server.Mount(router)
}
