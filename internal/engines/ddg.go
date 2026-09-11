package engines

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/FacileStudio/sonar/internal/httpc"
)

// DDG scrapes DuckDuckGo's html endpoint. It is the most hostile scraped
// engine from datacenter IPs (routinely returns a 202 anti-bot page), so it
// sits lowest in priority and is skipped entirely once tripped. Best-effort.
type DDG struct {
	Client *http.Client
	Count  int
}

func (d *DDG) Name() string { return "ddg" }

var reDDGResult = regexp.MustCompile(`(?s)<a[^>]+class="result__a"[^>]*href="([^"]+)"[^>]*>(.*?)</a>`)

func (d *DDG) Search(ctx context.Context, query string, count int) ([]Result, error) {
	form := url.Values{"q": {query}}
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://html.duckduckgo.com/html/", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpc.Fingerprint(req.Header)
	res, err := d.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTransient, err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusTooManyRequests || res.StatusCode == 202 {
		return nil, ErrBlocked
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d", ErrTransient, res.StatusCode)
	}
	body, _ := io.ReadAll(res.Body)
	return parseDDG(string(body)), nil
}

func parseDDG(page string) []Result {
	out := []Result{}
	for _, m := range reDDGResult.FindAllStringSubmatch(page, -1) {
		title := html2text(m[2])
		if title == "" {
			continue
		}
		out = append(out, Result{Title: title, URL: html2text(m[1]), Snippet: ddgSnippet(m[0]), Engine: "ddg"})
	}
	return out
}

func ddgSnippet(block string) string {
	re := regexp.MustCompile(`(?s)class="result__snippet"[^>]*>(.*?)</a>`)
	m := re.FindStringSubmatch(block)
	if m == nil {
		return ""
	}
	return html2text(m[1])
}
