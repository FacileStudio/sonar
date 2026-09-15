// Package mcpserver exposes sonar's search to an agent as MCP tools, spoken
// over stdio. Stdio is the transport for a local CLI: the engines read their
// keys and config from this machine, and a human has to wire this server into
// their own agent config, so there is no multi-user surface to expose.
package mcpserver

import (
	"context"

	"github.com/FacileStudio/sonar/internal/config"
	"github.com/FacileStudio/sonar/internal/search"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// serverName is how a client, and the human reading its permission prompt,
// refer to this server.
const serverName = "sonar"

// server holds the dispatcher for the life of the MCP process so per-engine
// circuit breakers and throttle warm-up persist across tool calls instead of
// resetting on every query.
type server struct {
	dispatcher *search.Dispatcher
}

// newServer builds a server with a dispatcher built once from config. It
// fails soft: a broken config or cache directory falls back to the per-call
// path rather than refusing to start.
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

// New builds the server with its search tool bound to an in-process call into
// internal/search, so the MCP surface and the `sonar search` command share
// the exact same dispatch, dedupe and ranking path.
//
// In process rather than shelling out to the sonar binary: a subprocess puts
// exit codes, coloured output and a second config load between the model and
// the results this package can already compute.
func New(version string) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: serverName, Version: version}, nil)
	if srv, ok := newServer(); ok {
		mcp.AddTool(s, searchTool(), func(ctx context.Context, req *mcp.CallToolRequest, in searchInput) (*mcp.CallToolResult, searchOutput, error) {
			return runSearchWith(ctx, req, in, srv.dispatcher)
		})
	} else {
		mcp.AddTool(s, searchTool(), func(ctx context.Context, req *mcp.CallToolRequest, in searchInput) (*mcp.CallToolResult, searchOutput, error) {
			return runSearchWith(ctx, req, in, nil)
		})
	}
	return s
}

// Serve runs the server over stdio until the client hangs up or ctx is done.
//
// Nothing reachable from here may write to stdout: it carries the JSON-RPC
// frames, and one stray line desynchronises the stream.
func Serve(ctx context.Context, version string) error {
	return New(version).Run(ctx, &mcp.StdioTransport{})
}
