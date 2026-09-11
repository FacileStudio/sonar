// splice.go holds the line-level helpers behind the nacelle config merge. The
// merge is a text splice: it works on the file as lines, so the lines sonar
// owns are the only ones that change and everything else survives byte for
// byte. mcpServerName and mcpSubcommand come from install.go, same package.
package mcpserver

import (
	"strings"
)

// entryLines returns the three lines sonar's server entry is made of, at
// serverIndent spaces for the name.
func entryLines(serverIndent int) []string {
	f := serverIndent + 2
	return []string{
		strings.Repeat(" ", serverIndent) + mcpServerName + ":",
		strings.Repeat(" ", f) + "command: " + mcpServerName,
		strings.Repeat(" ", f) + "args: [" + mcpSubcommand + "]",
	}
}

// blockEnd is the first line at or after baseLine whose non-blank, non-comment
// text sits at or shallower than baseIndent — i.e. where the block ends and the
// next sibling or outer key begins. It is the exclusive end of the range.
func blockEnd(lines []string, baseLine, baseIndent int) int {
	for i := baseLine + 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "" {
			continue
		}
		if indent0(lines[i]) <= baseIndent && strings.HasPrefix(strings.TrimSpace(lines[i]), "#") {
			continue
		}
		if indent0(lines[i]) <= baseIndent {
			return i
		}
	}
	return len(lines)
}

// nextServer is the first server key line (at exactly serverIndent) at or after
// start but still inside the block; it is the exclusive end of one entry.
func nextServer(lines []string, start, serverIndent, blockEnd int) int {
	for i := start; i < blockEnd; i++ {
		if strings.TrimSpace(lines[i]) == "" || strings.HasPrefix(strings.TrimSpace(lines[i]), "#") {
			continue
		}
		if indent0(lines[i]) == serverIndent {
			return i
		}
	}
	return blockEnd
}

// appendIndex is where a fresh entry goes inside the block: just after its last
// content line, before any trailing blank or comment lines that belong to it.
func appendIndex(lines []string, start, end int) int {
	last := start
	for i := start; i < end; i++ {
		if strings.TrimSpace(lines[i]) != "" && !strings.HasPrefix(strings.TrimSpace(lines[i]), "#") {
			last = i + 1
		}
	}
	return last
}

func indent0(line string) int {
	n := 0
	for n < len(line) && line[n] == ' ' {
		n++
	}
	return n
}

func concat(parts ...[]string) []string {
	var out []string
	for _, p := range parts {
		for _, s := range p {
			out = append(out, s)
		}
	}
	return out
}
