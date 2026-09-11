// The merge that installs sonar's MCP server declaration into a harness config.
// Ported from mycelium's internal/adapter/mcp.go, whose behaviour is proven
// there and whose tests sonar's tests mirror shape for shape.
package mcpserver

import (
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"reflect"
)

// serverCommand returns the executable an existing entry spawns, across both
// shapes: a bare command string, and an argv array whose first element is the
// binary.
func serverCommand(entry map[string]any) string {
	switch command := entry["command"].(type) {
	case string:
		return command
	case []any:
		if len(command) > 0 {
			first, _ := command[0].(string)
			return first
		}
	}
	return ""
}

// isSonarServer reports whether an existing entry was written by this tool, by
// key or by the binary it spawns, so an entry renamed by hand is still ours to
// replace rather than joined.
func isSonarServer(name string, entry map[string]any) bool {
	if name == mcpServerName {
		return true
	}
	return filepath.Base(serverCommand(entry)) == mcpServerName
}

// merge returns the full contents path should have with sonar declared under
// key, whether that differs from what is there, and whether there was anything
// safe to write at all.
//
// Merge rather than overwrite, because every one of these files also holds
// servers and settings sonar did not put there. Replace rather than append,
// because the entry sonar owns is sonar's whatever it happens to be called.
//
// Unparseable JSON is refused outright, matching writeConfig: a file we cannot
// read is a file whose contents we would be guessing at, and this output is
// written over the original.
//
// The result is deterministic — encoding/json sorts map keys — so a second
// install re-reads what the first wrote and produces the same bytes.
func merge(path, key string, server map[string]any) (string, bool, bool) {
	settings := make(map[string]any)
	if data, err := os.ReadFile(path); err == nil {
		if json.Unmarshal(data, &settings) != nil {
			return "", false, false
		}
	}

	servers, _ := settings[key].(map[string]any)
	if servers == nil {
		servers = make(map[string]any)
	}
	before := maps.Clone(servers)

	for name, entry := range servers {
		existing, _ := entry.(map[string]any)
		if isSonarServer(name, existing) {
			delete(servers, name)
		}
	}
	servers[mcpServerName] = server
	settings[key] = servers

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return "", false, false
	}
	return string(data) + "\n", !reflect.DeepEqual(before, servers), true
}
