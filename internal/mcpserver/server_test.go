package mcpserver

import (
	"context"
	"reflect"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// connect wires a client to the server over in-memory transports, which
// exercises the same code path stdio does without touching a real stdin.
func connect(t *testing.T) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	serverSide, clientSide := mcp.NewInMemoryTransports()
	ss, err := New("test").Connect(ctx, serverSide, nil)
	if err != nil {
		t.Fatalf("connecting the server: %v", err)
	}
	t.Cleanup(func() { ss.Close() })
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "test"}, nil)
	cs, err := client.Connect(ctx, clientSide, nil)
	if err != nil {
		t.Fatalf("connecting the client: %v", err)
	}
	t.Cleanup(func() { cs.Close() })
	return cs
}

// listTools returns what tools/list reports, in the order it reported it.
func listTools(t *testing.T) []*mcp.Tool {
	t.Helper()
	res, err := connect(t).ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("tools/list: %v", err)
	}
	return res.Tools
}

// The server's one purpose is the search tool; a second one appearing here is
// a feature somebody should have decided on, not a surprise from the wiring.
func TestToolsListReportsExactlySearch(t *testing.T) {
	tools := listTools(t)
	if len(tools) != 1 || tools[0].Name != "search" {
		t.Fatalf("tools = %+v, want exactly one tool named search", tools)
	}
}

// Annotations drive the client's permission UI. Every field is pinned because
// three of the four default to the permissive value when left unset, so "we
// forgot" and "we meant it" look identical on the wire. Search is read-only and
// open-world: it never changes local state, but it does reach the public web.
func TestSearchToolSetsAllFourAnnotationsExplicitly(t *testing.T) {
	want := &mcp.ToolAnnotations{
		ReadOnlyHint: true, DestructiveHint: new(false), IdempotentHint: true, OpenWorldHint: new(true),
	}
	if !reflect.DeepEqual(listTools(t)[0].Annotations, want) {
		t.Fatalf("search annotations = %+v, want %+v", listTools(t)[0].Annotations, want)
	}
}

// A tool with no input schema cannot be called, and one with no output schema
// hands the model prose to re-parse instead of data. Both are declared by the
// argument and result types, so this catches a type that stopped generating one.
func TestSearchToolAdvertisesAnObjectInputAndOutputSchema(t *testing.T) {
	for label, schema := range map[string]any{"inputSchema": listTools(t)[0].InputSchema, "outputSchema": listTools(t)[0].OutputSchema} {
		if schema == nil {
			t.Errorf("search has no %s", label)
			continue
		}
		object, ok := schema.(map[string]any)
		if !ok || object["type"] != "object" {
			t.Errorf("search %s = %v, want a JSON schema of type object", label, schema)
		}
	}
}
