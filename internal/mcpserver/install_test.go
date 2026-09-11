// install_test covers the merge and declare path for wiring sonar's MCP server
// into harness configs. The merge is ported from mycelium's internal/adapter
// mcp_test.go, so these mirror its cases: shape, idempotence, foreign servers
// surviving, refusal on unparseable JSON, and the OpenCode local shape.
package mcpserver

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// agentAt builds an Agent whose config lives in a temp dir under t, so a test
// can hand a fake home to one harness without touching the real machine.
func agentAt(t *testing.T, name, key, config string) Agent {
	t.Helper()
	dir := filepath.Join(t.TempDir(), name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	return Agent{Name: name, Path: filepath.Join(dir, config), Key: key,
		Marker: filepath.Join(dir, "marker")}
}

func serversIn(t *testing.T, content, key string) map[string]any {
	t.Helper()
	var settings map[string]any
	if err := json.Unmarshal([]byte(content), &settings); err != nil {
		t.Fatalf("merged content is not JSON: %v\n%s", err, content)
	}
	servers, ok := settings[key].(map[string]any)
	if !ok {
		t.Fatalf("no %q object in:\n%s", key, content)
	}
	return servers
}

// TestTheDeclarationNamesTheBinaryAndTheStdioSubcommand is the shape check: a
// client that reads this file must end up spawning `<binary> mcp`.
func TestTheDeclarationNamesTheBinaryAndTheStdioSubcommand(t *testing.T) {
	a := agentAt(t, "claude", "mcpServers", "config.json")
	content, _, ok := merge(a.Path, a.Key, serverFor(a))
	if !ok {
		t.Fatal("expected a declaration for a fresh install")
	}
	entry, isMap := serversIn(t, content, a.Key)[mcpServerName].(map[string]any)
	if !isMap {
		t.Fatalf("no %q server in:\n%s", mcpServerName, content)
	}
	if command, _ := entry["command"].(string); command == "" {
		t.Errorf("entry has no command: %v", entry)
	}
	args, _ := entry["args"].([]any)
	if len(args) != 1 || args[0] != mcpSubcommand {
		t.Errorf("args = %v, want [%q]", args, mcpSubcommand)
	}
}

// TestOpencodeGetsTheLocalTransportShape pins OpenCode's divergence: a tagged
// transport and one argv array, no separate args key.
func TestOpencodeGetsTheLocalTransportShape(t *testing.T) {
	a := agentAt(t, "opencode", "mcp", "config.json")
	content, _, _ := merge(a.Path, a.Key, serverFor(a))
	entry, _ := serversIn(t, content, a.Key)[mcpServerName].(map[string]any)
	if entry == nil || entry["type"] != "local" {
		t.Errorf("opencode entry = %v, want a type:local transport", entry)
	}
	if command, _ := entry["command"].([]any); len(command) != 2 || command[1] != mcpSubcommand {
		t.Errorf("opencode command = %v, want [<binary>, %q]", entry["command"], mcpSubcommand)
	}
	if _, hasArgs := entry["args"]; hasArgs {
		t.Errorf("opencode must not get an args key: %v", entry)
	}
}

// TestASecondMergeIsByteIdenticalToTheFirst is the property the installer's
// idempotence depends on: merge reads what the last install wrote, so anything
// non-deterministic shows up forever as a harness that is permanently rewiring.
func TestASecondMergeIsByteIdenticalToTheFirst(t *testing.T) {
	a := agentAt(t, "claude", "mcpServers", "config.json")
	writeFixture(t, a.Path, `{"numTurns":3,"mcpServers":{"figma":{"command":"figma-mcp"}}}`)

	first, _, ok := merge(a.Path, a.Key, serverFor(a))
	if !ok {
		t.Fatal("first merge produced nothing")
	}
	if err := os.WriteFile(a.Path, []byte(first), 0o644); err != nil {
		t.Fatal(err)
	}
	second, _, _ := merge(a.Path, a.Key, serverFor(a))
	if second != first {
		t.Errorf("second merge differs from the first:\n--- first\n%s\n--- second\n%s", first, second)
	}
}

// TestForeignServersAndUnknownKeysSurviveTheMerge is why this reads the file
// instead of writing a fresh one: these files hold servers and settings sonar
// did not put there, and install must not overwrite them.
func TestForeignServersAndUnknownKeysSurviveTheMerge(t *testing.T) {
	a := agentAt(t, "claude", "mcpServers", "config.json")
	writeFixture(t, a.Path, `{"theme":"dark","mcpServers":{"figma":{"command":"figma-mcp","args":["--stdio"]}}}`)

	content, _, _ := merge(a.Path, a.Key, serverFor(a))
	var settings map[string]any
	if err := json.Unmarshal([]byte(content), &settings); err != nil {
		t.Fatal(err)
	}
	if settings["theme"] != "dark" {
		t.Errorf("unknown top-level key lost:\n%s", content)
	}
	figma, _ := serversIn(t, content, a.Key)["figma"].(map[string]any)
	if figma == nil || figma["command"] != "figma-mcp" {
		t.Errorf("foreign server lost or rewritten:\n%s", content)
	}
}

// TestAnEntryRenamedByHandIsStillOurs covers the half of the stale-entry bug
// that matching the key misses: whoever renamed the entry did not change what
// it runs, so it is recognised by the binary it spawns.
func TestAnEntryRenamedByHandIsStillOurs(t *testing.T) {
	a := agentAt(t, "claude", "mcpServers", "config.json")
	writeFixture(t, a.Path, `{"mcpServers":{"search":{"command":"/opt/old/sonar","args":["mcp"]}}}`)

	content, _, _ := merge(a.Path, a.Key, serverFor(a))
	if _, stale := serversIn(t, content, a.Key)["search"]; stale {
		t.Errorf("a renamed sonar entry survived:\n%s", content)
	}
}

// TestAnUnreadableConfigIsLeftAlone matches merge's refusal: an unparseable
// file is a file whose contents an install would be guessing at, so nothing is
// written and Declare names it.
func TestAnUnreadableConfigIsLeftAlone(t *testing.T) {
	a := agentAt(t, "claude", "mcpServers", "config.json")
	writeFixture(t, a.Path, "{not json")

	content, _, ok := merge(a.Path, a.Key, serverFor(a))
	if ok {
		t.Fatalf("expected a refusal on corrupt JSON, got:\n%s", content)
	}
	changed, err := Declare(a)
	if err == nil || changed {
		t.Errorf("Declare claims success on an unparseable config: changed=%v err=%v", changed, err)
	}
	if data, _ := os.ReadFile(a.Path); string(data) != "{not json" {
		t.Errorf("corrupt config was rewritten: %q", data)
	}
}

// TestOpencodeJsoncWinsLeavesTheJsonSiblingAlone: when OpenCode could read
// either file, sonar cannot know which, so it writes neither.
func TestOpencodeJsoncWinsLeavesTheJsonSiblingAlone(t *testing.T) {
	dir := t.TempDir()
	a := Agent{Name: "opencode", Path: filepath.Join(dir, "opencode.json"),
		Key: "mcp", Marker: filepath.Join(dir, "marker")}
	jsonPath := a.Path
	writeFixture(t, jsonPath, `{}`)
	if err := os.WriteFile(filepath.Join(dir, "opencode.jsonc"), []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Declare(a); err == nil {
		t.Error("Declare should refuse when opencode.jsonc is present")
	}
	if data, _ := os.ReadFile(jsonPath); string(data) != `{}` {
		t.Errorf("the .json sibling was rewritten: %q", data)
	}
}

// TestDeclareWritesAndIsIdempotent exercises the full declare path once and
// then again, asserting the second run changes nothing.
func TestDeclareWritesAndIsIdempotent(t *testing.T) {
	a := agentAt(t, "claude", "mcpServers", "config.json")
	changed, err := Declare(a)
	if err != nil || !changed {
		t.Fatalf("first declare: changed=%v err=%v", changed, err)
	}
	first, _ := os.ReadFile(a.Path)

	changed, err = Declare(a)
	if err != nil || changed {
		t.Fatalf("second declare should be a no-op: changed=%v err=%v", changed, err)
	}
	second, _ := os.ReadFile(a.Path)
	if string(first) != string(second) {
		t.Errorf("idempotent declare rewrote the file:\n%s", second)
	}
}

func writeFixture(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
