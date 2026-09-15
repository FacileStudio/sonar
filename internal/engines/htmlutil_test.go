package engines

import (
	"strings"
	"testing"
)

func TestHTMLToMarkdown(t *testing.T) {
	input := `<html>
<head><title>Test Page</title><style>body { color: red; }</style></head>
<body>
<nav><a href="/home">Home</a></nav>
<script>var x = 123;</script>
<h1>Welcome to Sonar</h1>
<p>Sonar is a <strong>multi-engine</strong> search tool with <a href="https://example.com">links</a>.</p>
<table>
<thead><tr><th>Header 1</th><th>Header 2</th></tr></thead>
<tbody><tr><td>Cell 1</td><td>Cell 2</td></tr></tbody>
</table>
</body>
</html>`

	md, err := HTMLToMarkdown(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.Contains(md, "var x = 123") {
		t.Errorf("expected script content to be removed, got %q", md)
	}
	if strings.Contains(md, "color: red") {
		t.Errorf("expected style content to be removed, got %q", md)
	}
	if !strings.Contains(md, "# Welcome to Sonar") {
		t.Errorf("expected header markdown, got %q", md)
	}
	if !strings.Contains(md, "[links](https://example.com)") {
		t.Errorf("expected link markdown, got %q", md)
	}
}
