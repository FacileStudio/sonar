package engines

import (
	"html"
	"regexp"
	"strings"
)

var reTag = regexp.MustCompile(`(?s)<[^>]+>`)

// html2text strips HTML tags and unescapes entities from a scrape fragment.
func html2text(s string) string {
	s = reTag.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	return strings.TrimSpace(s)
}
