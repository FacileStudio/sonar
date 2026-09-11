// Package mcpserver exposes sonar's search to an agent as MCP tools, spoken
// over stdio. Stdio is the transport for a local CLI: the engines read their
// keys and config from this machine, and a human has to wire this server into
// their own agent config, so there is no multi-user surface to expose.
package mcpserver

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// serverName is how a client, and the human reading its permission prompt,
// refer to this server.
const serverName = "sonar"

// New builds the server with its search tool bound to an in-process call into
// internal/search, so the MCP surface and the `sonar search` command share
// the exact same dispatch, dedupe and ranking path.
//
// In process rather than shelling out to the sonar binary: a subprocess puts
// exit codes, coloured output and a second config load between the model and
// the results this package can already compute.
func New(version string) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: serverName, Version: version}, nil)
	mcp.AddTool(s, searchTool(), runSearch)
	return s
}

// Serve runs the server over stdio until the client hangs up or ctx is done.
//
// Nothing reachable from here may write to stdout: it carries the JSON-RPC
// frames, and one stray line desynchronises the stream.
func Serve(ctx context.Context, version string) error {
	return New(version).Run(ctx, &mcp.StdioTransport{})
}

// hint returns a pointer to a bool, which the pointer-valued annotations need
// so the SDK can tell false from unset.
//
// DestructiveHint and OpenWorldHint show as true when absent, so a tool that
// omits them is advertised as destructive and open-world. For a read-only
// search that is the exact opposite of the truth.
func hint(b bool) *bool { return &b }
