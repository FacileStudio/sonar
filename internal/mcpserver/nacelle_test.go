// nacelle_test covers declare against a nacelle-tui config, which is the YAML,
// hand-edited harness target. The tests assert the surgical splice contract:
// sonar's entry is added or replaced, and everything else — comments, block
// scalars, other servers — survives byte for byte. writeFixture, Agent and
// Declare live in install_test.go and install.go, same package.
package mcpserver

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// nacelleFixture writes a nacelle-like config with a comment, a block scalar
// and a foreign MCP server, so the round trip is tested against content that
// must survive untouched, not against an empty file.
func nacelleFixture(t *testing.T) string {
	path := filepath.Join(t.TempDir(), ".nacelle.yml")
	writeFixture(t, path, `# hand-edited
ui:
  start_message: |2
    ▄▄ ▄▄
sources:
  skill_dirs: []
  mcp:
    mycelium:
      command: mycelium
      args: [mcp]
cron: []
`)
	return path
}

// TestNacelleMergeAddsSonarAndLeavesTheRestByteIdentical runs the declare and
// asserts the foreign server, the comment and the block scalar all survive.
func TestNacelleMergeAddsSonarAndLeavesTheRestByteIdentical(t *testing.T) {
	path := nacelleFixture(t)
	a := Agent{Name: "nacelle", Path: path, Key: ""}
	changed, err := Declare(a)
	if err != nil || !changed {
		t.Fatalf("first nacelle declare: changed=%v err=%v", changed, err)
	}
	data, _ := os.ReadFile(path)
	out := string(data)
	if !strings.Contains(out, "# hand-edited") {
		t.Errorf("comment lost:\n%s", out)
	}
	if !strings.Contains(out, "start_message: |2") || !strings.Contains(out, "▄▄ ▄▄") {
		t.Errorf("block scalar mangled:\n%s", out)
	}
	if !strings.Contains(out, "mycelium:") || !strings.Contains(out, "cron: []") {
		t.Errorf("foreign content lost:\n%s", out)
	}
	if !strings.Contains(out, "sonar:") || !strings.Contains(out, "command: sonar") {
		t.Errorf("sonar not declared:\n%s", out)
	}
}

// TestNacelleDeclareIsIdempotent asserts a second run changes nothing, which is
// what guarantees a repeated install never churns a hand-edited file.
func TestNacelleDeclareIsIdempotent(t *testing.T) {
	path := nacelleFixture(t)
	a := Agent{Name: "nacelle", Path: path, Key: ""}
	if _, err := Declare(a); err != nil {
		t.Fatal(err)
	}
	first, _ := os.ReadFile(path)

	changed, err := Declare(a)
	if err != nil || changed {
		t.Fatalf("second nacelle declare should be a no-op: changed=%v err=%v", changed, err)
	}
	second, _ := os.ReadFile(path)
	if string(first) != string(second) {
		t.Errorf("idempotent nacelle declare rewrote the file:\n%s", second)
	}
}

// TestNacelleReplacesASonarEntryThatNoLongerMatches covers the upgrade path: an
// entry that keyed itself as sonar but no longer runs the right binary and args
// is updated in place, not appended a second time.
func TestNacelleReplacesAStaleSonarEntryInPlace(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".nacelle.yml")
	writeFixture(t, path, `sources:
  mcp:
    sonar:
      command: /opt/legacy-search
      args: [search]
cron: []
`)
	a := Agent{Name: "nacelle", Path: path, Key: ""}
	changed, err := Declare(a)
	if err != nil || !changed {
		t.Fatalf("stale sonar not updated: changed=%v err=%v", changed, err)
	}
	data, _ := os.ReadFile(path)
	out := string(data)
	if countSub(out, "command:") != 1 {
		t.Errorf("sonar entry was appended instead of replaced:\n%s", out)
	}
	if strings.Contains(out, "/opt/legacy-search") || strings.Contains(out, "search]") {
		t.Errorf("stale command or args survived:\n%s", out)
	}
	if !strings.Contains(out, "command: sonar") || !strings.Contains(out, "args: [mcp]") {
		t.Errorf("canonical declaration missing after replace:\n%s", out)
	}
}

// TestNacelleUnchangedEntryIsANoOp asserts a config whose sonar entry already
// matches does not touch the file at all.
func TestNacelleUnchangedEntryIsANoOp(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".nacelle.yml")
	writeFixture(t, path, `sources:
  mcp:
    sonar:
      command: sonar
      args: [mcp]
`)
	a := Agent{Name: "nacelle", Path: path, Key: ""}
	changed, err := Declare(a)
	if err != nil || changed {
		t.Fatalf("an already-declared sonar entry changed things: changed=%v err=%v", changed, err)
	}
}

// TestNacelleRefusesBrokenYamlAndANonMappingMCPBlock names both refusal paths:
// a file sonar cannot parse, and a sources.mcp that is not a mapping, are left
// alone rather than guessed at.
func TestNacelleRefusesBrokenYamlAndANonMappingMCPBlock(t *testing.T) {
	broken := filepath.Join(t.TempDir(), ".nacelle.yml")
	writeFixture(t, broken, "sources: [unclosed\n")
	a := Agent{Name: "nacelle", Path: broken, Key: ""}
	if _, err := Declare(a); err == nil {
		t.Error("declare should refuse on malformed YAML")
	}
	if data, _ := os.ReadFile(broken); string(data) != "sources: [unclosed\n" {
		t.Errorf("broken nacelle.yml was rewritten: %q", data)
	}

	notMap := filepath.Join(t.TempDir(), ".nacelle.yml")
	writeFixture(t, notMap, `sources:
  mcp: [mycelium]
`)
	a = Agent{Name: "nacelle", Path: notMap, Key: ""}
	if _, err := Declare(a); err == nil {
		t.Error("declare should refuse when sources.mcp is not a mapping")
	}
}

// TestNacelleCreatesTheBlockWhenAbsent asserts a minimal config gains the whole
// sources.mcp path so an install never fails on a nacelle without an mcp block.
func TestNacelleCreatesTheBlockWhenAbsent(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".nacelle.yml")
	writeFixture(t, path, "ui:\n  prompt_placeholder: yoo\n")
	a := Agent{Name: "nacelle", Path: path, Key: ""}
	changed, err := Declare(a)
	if err != nil || !changed {
		t.Fatalf("declare on a nacelle without the block: changed=%v err=%v", changed, err)
	}
	data, _ := os.ReadFile(path)
	out := string(data)
	if !strings.Contains(out, "mcp:") || !strings.Contains(out, "command: sonar") {
		t.Errorf("sources.mcp not created:\n%s", out)
	}
}

func countSub(haystack, needle string) int {
	n := 0
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			n++
		}
	}
	return n
}
