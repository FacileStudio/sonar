package server

import (
	"context"
	"log/slog"

	"github.com/FacileStudio/porte/mcpauth"
	"github.com/FacileStudio/sonar/internal/server/modules/mcp"
	mcpkit "github.com/FacileStudio/tronc/mcp"
)

// newMCP builds the MCP server when sign-in is on and MCP_RESOURCE is set,
// and nothing otherwise. The verifier runs Registre discovery here, so a
// wrong issuer fails at boot like sign-in does. A person's first MCP call
// creates their account through the web login's own path, so Bulle and other
// clients reach sonar for people who never opened it.
func newMCP(ctx context.Context, a *app, cfg Config, version string) (*mcp.Service, error) {
	if cfg.OIDC.Issuer == "" || cfg.MCPResource == "" {
		return nil, nil
	}
	verifier, err := mcpauth.Verifier(ctx, mcpauth.Config{Issuer: cfg.OIDC.Issuer, Resource: cfg.MCPResource,
		ScopePrefix: "sonar", Identities: a.kits.Identities, Provision: true, Accounts: a.kits.OIDC, Logger: a.log})
	if err != nil {
		return nil, err
	}
	a.log.Info("mcp server on", slog.String("resource", cfg.MCPResource))
	return mcp.NewService(a.search, mcpkit.AppConfig{Resource: cfg.MCPResource, Issuer: cfg.OIDC.Issuer, Version: version,
		Origins: cfg.CORSAllowedOrigins, Verifier: verifier, Logger: a.log})
}
