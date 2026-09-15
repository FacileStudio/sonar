// Package mcpserver exposes sonar's search and fetch tools to an agent as MCP tools,
// spoken over stdio.
package mcpserver

import (
	"context"

	"github.com/FacileStudio/sonar/internal/config"
	"github.com/FacileStudio/sonar/internal/search"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const serverName = "sonar"

type server struct {
	dispatcher *search.Dispatcher
}

func newServer() (*server, bool) {
	cfg, err := config.Load(config.Path())
	if err != nil {
		return nil, false
	}
	d, err := search.NewDispatcher(cfg)
	if err != nil {
		return nil, false
	}
	return &server{dispatcher: d}, true
}

func registerTools(s *mcp.Server, d *search.Dispatcher) {
	mcp.AddTool(s, searchTool(), func(ctx context.Context, req *mcp.CallToolRequest, in searchInput) (*mcp.CallToolResult, searchOutput, error) {
		return runSearchWith(ctx, req, in, d)
	})
	mcp.AddTool(s, fetchTool(), runFetch)
}

// New builds the server with search and fetch tools bound to in-process handlers.
func New(version string) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: serverName, Version: version}, nil)
	if srv, ok := newServer(); ok {
		registerTools(s, srv.dispatcher)
	} else {
		registerTools(s, nil)
	}
	return s
}

// Serve runs the server over stdio until the client hangs up or ctx is done.
func Serve(ctx context.Context, version string) error {
	return New(version).Run(ctx, &mcp.StdioTransport{})
}
