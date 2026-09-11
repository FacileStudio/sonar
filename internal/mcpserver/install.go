// The install side of the MCP server: registering the search tool in the agent
// harnesses on this machine, so a harness whose config is a JSON file spawns
// `sonar mcp` from the declaration installed here instead of the model shelling
// out to the binary.
package mcpserver

import (
	"fmt"
	"os"
	"path/filepath"
)

// mcpServerName is the key sonar declares itself under, and mcpSubcommand the
// argument that puts the binary on stdio. Both mirrored here so a rename cannot
// leave a harness pointed at a command that is gone.
const mcpServerName = "sonar"
const mcpSubcommand = "mcp"

// Agent is one harness sonar can declare itself in and where its config lives.
// Marker is the harness's own directory, the proof it has run here once.
// Codex is absent on purpose: its MCP servers live in TOML, and sonar — like
// mycelium — does not write TOML.
type Agent struct {
	Name   string
	Path   string
	Key    string
	Marker string
}

// mcpBinary returns the running binary's own absolute path. A harness spawns
// its servers directly rather than through a login shell, so ~/.local/bin is
// routinely off the PATH it inherits; the binary you ran install with is the
// one the config should point at.
func mcpBinary() string {
	exe, err := os.Executable()
	if err != nil {
		return mcpServerName
	}
	return exe
}

// serverFor returns the declaration in the shape the agent's config reads:
// OpenCode tags the transport and takes one argv array with no args key, the
// others take a command plus args.
func serverFor(a Agent) map[string]any {
	if a.Name == "opencode" {
		return map[string]any{
			"type":    "local",
			"command": []any{mcpBinary(), mcpSubcommand},
		}
	}
	return map[string]any{
		"command": mcpBinary(),
		"args":    []any{mcpSubcommand},
	}
}

// atomic reports whether the agent's config is written by temp-file rename.
// ~/.claude.json is Claude Code's live state store — session history, the
// account record — not a generated file, so it must never sit truncated.
func atomic(a Agent) bool {
	return a.Name == "claude"
}

// opencodeHasJsonc reports whether a .jsonc sibling exists. OpenCode accepts
// the comments and trailing commas in .jsonc that encoding/json refuses, and a
// .jsonc present means we cannot know which file OpenCode reads, so sonar
// writes nothing and leaves the .json form alone.
func opencodeHasJsonc(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "opencode.jsonc"))
	return err == nil
}

// writeConfig replaces the agent's config with content. ~/.claude.json is
// Claude Code's live state store, so only it gets the atomic temp-file path.
func writeConfig(a Agent, content string) error {
	if atomic(a) {
		return writeAtomic(a.Path, content)
	}
	if err := os.MkdirAll(filepath.Dir(a.Path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(a.Path, []byte(content), 0o644)
}

// writeAtomic replaces a file without ever leaving it truncated, and pins it
// 0600 because ~/.claude.json carries the account record.
//
// The temp file sits in the same directory so the rename cannot cross a
// filesystem, and 0600 is set explicitly because os.CreateTemp's mode is not
// something a reader of this function should have to go and check.
func writeAtomic(path, content string) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), filepath.Base(path)+".*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()

	if _, err := tmp.WriteString(content); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// Declare merges sonar's server into the agent's config and writes it back,
// reporting whether the declaration changed. A file sonar cannot safely touch —
// unparseable JSON, or an opencode.jsonc sibling — is left alone and returned
// as an error so the caller can name it instead of guessing.
func Declare(a Agent) (bool, error) {
	if a.Name == "nacelle" {
		return declareNacelle(a.Path)
	}
	if a.Name == "opencode" && opencodeHasJsonc(filepath.Dir(a.Path)) {
		return false, fmt.Errorf("opencode.jsonc wins over opencode.json; nothing written")
	}
	content, changed, ok := merge(a.Path, a.Key, serverFor(a))
	if !ok {
		return false, fmt.Errorf("%s is not valid JSON; left alone", a.Path)
	}
	if !changed {
		return false, nil
	}
	if err := writeConfig(a, content); err != nil {
		return false, err
	}
	return true, nil
}
