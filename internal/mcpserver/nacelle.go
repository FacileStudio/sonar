// nacelle.go merges sonar's MCP server into the sources.mcp block of a
// nacelle-tui config, which is the one harness target that is YAML and
// hand-edited rather than generated JSON. The merge is surgical: it parses the
// file only to locate the block and its servers, then splices text so the
// lines sonar owns are the only ones that change. Everything else — comments,
// blank lines, block scalars, other servers — survives byte for byte. A whole
// file round trip through the yaml emitter is deliberately avoided because it
// re-indents and reflows content that the user did not ask to touch.
package mcpserver

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// declareNacelle merges sonar's entry into the sources.mcp block of the config
// at path, reporting whether the file changed. The declaration is canonical
// `command: sonar` with `args: [mcp]` — a bare name, not an absolute path,
// because nacelle resolves servers on PATH the way the existing mycelium entry
// does.
func declareNacelle(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return false, fmt.Errorf("%s is not valid YAML; left alone", path)
	}
	root := &doc
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		root = root.Content[0]
	}
	if root.Kind != yaml.MappingNode {
		return false, fmt.Errorf("%s does not start with a mapping; left alone", path)
	}

	newLines, changed, err := splice(strings.Split(string(data), "\n"), root)
	if err != nil {
		return false, err
	}
	if !changed {
		return false, nil
	}
	if err := os.WriteFile(path, []byte(strings.Join(newLines, "\n")), 0o644); err != nil {
		return false, err
	}
	return true, nil
}

// splice edits the sonar MCP entry into lines and returns the result. It is
// the only place that changes content; everything it does not touch is left
// exactly as it was.
func splice(lines []string, root *yaml.Node) ([]string, bool, error) {
	sk, sv, hasSources := mappingKey(root, "sources")
	if hasSources && sv.Kind != yaml.MappingNode && sv.Kind != 0 {
		return lines, false, fmt.Errorf("sources is not a mapping; left alone")
	}
	if !hasSources || sv.Kind != yaml.MappingNode {
		return addBlock(lines, 0, len(lines), []string{"", "sources:", strings.Repeat(" ", 2) + "mcp:"}, 4), true, nil
	}
	mk, mv, hasMCP := mappingKey(sv, "mcp")
	if hasMCP && mv.Kind != yaml.MappingNode && mv.Kind != 0 {
		return lines, false, fmt.Errorf("sources.mcp is not a mapping; left alone")
	}
	if hasMCP && mv.Kind == yaml.MappingNode {
		return editServers(lines, mk, mv)
	}
	si := sk.Column - 1
	return addBlock(lines, sk.Line-1, blockEnd(lines, sk.Line-1, si), []string{strings.Repeat(" ", si+2) + "mcp:"}, si+4), true, nil
}

// addBlock inserts prefix followed by sonar's canonical entry into lines, where
// a new child of the block starting at start belongs: just after the last
// content line, before any trailing blank or comment lines.
func addBlock(lines []string, start, end int, prefix []string, serverIndent int) []string {
	insert := appendIndex(lines, start, end)
	return concat(lines[0:insert], concat(prefix, entryLines(serverIndent)), lines[insert:])
}

// editServers updates or appends sonar within an existing sources.mcp mapping.
func editServers(lines []string, mcpKey, mcpValue *yaml.Node) ([]string, bool, error) {
	baseIndent := mcpKey.Column - 1
	end := blockEnd(lines, mcpKey.Line-1, baseIndent)

	var key, value *yaml.Node
	for i := 0; i+1 < len(mcpValue.Content); i += 2 {
		k, v := mcpValue.Content[i], mcpValue.Content[i+1]
		if sonarServerNode(k, v) {
			key, value = k, v
			break
		}
	}

	if isCanonical(value) {
		return lines, false, nil
	}
	if key != nil {
		start := key.Line - 1
		stop := nextServer(lines, start+1, baseIndent+2, end)
		return concat(lines[0:start], entryLines(baseIndent+2), lines[stop:]), true, nil
	}
	insert := appendIndex(lines, mcpKey.Line-1, end)
	return concat(lines[0:insert], entryLines(baseIndent+2), lines[insert:]), true, nil
}

// sonarServerNode recognises a server entry either by its name key or by the
// command it runs, so an entry renamed by hand is still understood as sonar's.
func sonarServerNode(k, v *yaml.Node) bool {
	if k.Kind == yaml.ScalarNode && k.Value == mcpServerName {
		return true
	}
	return commandRuns(v, mcpServerName)
}

// isCanonical reports whether the server already carries the exact declaration
// sonar would write, which is what makes a repeated install a no-op.
func isCanonical(v *yaml.Node) bool {
	if v == nil || v.Kind != yaml.MappingNode {
		return false
	}
	if !commandRuns(v, mcpServerName) {
		return false
	}
	_, args, ok := mappingKey(v, "args")
	if !ok || args.Kind != yaml.SequenceNode {
		return false
	}
	return len(args.Content) == 1 && args.Content[0].Kind == yaml.ScalarNode &&
		args.Content[0].Value == mcpSubcommand
}

// commandRuns reports whether the mapping node's command (a scalar, or the
// first element of a sequence) is the named binary, matched on base name the
// way the JSON harness merge does.
func commandRuns(v *yaml.Node, bin string) bool {
	if v == nil || v.Kind != yaml.MappingNode {
		return false
	}
	_, cv, ok := mappingKey(v, "command")
	if !ok {
		return false
	}
	if cv.Kind == yaml.ScalarNode {
		return filepath.Base(cv.Value) == bin
	}
	if cv.Kind == yaml.SequenceNode && len(cv.Content) > 0 && cv.Content[0].Kind == yaml.ScalarNode {
		return filepath.Base(cv.Content[0].Value) == bin
	}
	return false
}

// mappingKey finds the key node and value node for key in a mapping, reporting
// whether the key is present.
func mappingKey(mapping *yaml.Node, key string) (*yaml.Node, *yaml.Node, bool) {
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return nil, nil, false
	}
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		k, v := mapping.Content[i], mapping.Content[i+1]
		if k.Kind == yaml.ScalarNode && k.Value == key {
			return k, v, true
		}
	}
	return nil, nil, false
}
