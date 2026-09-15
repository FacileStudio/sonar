package mcpserver

import (
	"context"
	"reflect"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

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

func listTools(t *testing.T) []*mcp.Tool {
	t.Helper()
	res, err := connect(t).ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("tools/list: %v", err)
	}
	return res.Tools
}

func TestToolsListReportsSearchAndFetch(t *testing.T) {
	tools := listTools(t)
	if len(tools) != 2 {
		t.Fatalf("got %d tools, want exactly 2 (search, fetch)", len(tools))
	}
	names := map[string]bool{tools[0].Name: true, tools[1].Name: true}
	if !names["search"] || !names["fetch"] {
		t.Fatalf("tools = %+v, want search and fetch", tools)
	}
}

func TestToolAnnotationsExplicit(t *testing.T) {
	want := &mcp.ToolAnnotations{
		ReadOnlyHint: true, DestructiveHint: new(false), IdempotentHint: true, OpenWorldHint: new(true),
	}
	for _, tool := range listTools(t) {
		if !reflect.DeepEqual(tool.Annotations, want) {
			t.Fatalf("tool %s annotations = %+v, want %+v", tool.Name, tool.Annotations, want)
		}
	}
}

func TestToolSchemas(t *testing.T) {
	for _, tool := range listTools(t) {
		for label, schema := range map[string]any{"inputSchema": tool.InputSchema, "outputSchema": tool.OutputSchema} {
			if schema == nil {
				t.Errorf("%s has no %s", tool.Name, label)
				continue
			}
			object, ok := schema.(map[string]any)
			if !ok || object["type"] != "object" {
				t.Errorf("%s %s = %v, want object schema", tool.Name, label, schema)
			}
		}
	}
}
