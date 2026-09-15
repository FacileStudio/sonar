package engines

import (
	"html"
	"regexp"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
)

var (
	reTag      = regexp.MustCompile(`(?s)<[^>]+>`)
	reScript   = regexp.MustCompile(`(?is)<script\b[^>]*>.*?</script>`)
	reStyle    = regexp.MustCompile(`(?is)<style\b[^>]*>.*?</style>`)
	reNoscript = regexp.MustCompile(`(?is)<noscript\b[^>]*>.*?</noscript>`)
	reSvg      = regexp.MustCompile(`(?is)<svg\b[^>]*>.*?</svg>`)
)

// CleanHTML removes script, style, noscript, and svg blocks from HTML.
func CleanHTML(s string) string {
	s = reScript.ReplaceAllString(s, "")
	s = reStyle.ReplaceAllString(s, "")
	s = reNoscript.ReplaceAllString(s, "")
	s = reSvg.ReplaceAllString(s, "")
	return s
}

// HTMLToMarkdown converts raw HTML into GitHub Flavored Markdown.
func HTMLToMarkdown(raw string) (string, error) {
	cleaned := CleanHTML(raw)
	md, err := htmltomarkdown.ConvertString(cleaned)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(md), nil
}

// html2text strips HTML tags and unescapes entities from a scrape fragment.
func html2text(s string) string {
	s = reTag.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	return strings.TrimSpace(s)
}
