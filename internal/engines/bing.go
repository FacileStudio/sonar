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

// Bing scrapes the HTML result page. It is a last-resort engine: retention
// varies by IP reputation, and single-domain queries resolve reliably while
// multi-concept ones can return unrelated results. Treated as unreliable.
type Bing struct {
	Client *http.Client
	Count  int
}

func (b *Bing) Name() string { return "bing" }

var (
	reBingBlock = regexp.MustCompile(`<li class="b_algo"`)
	reH2        = regexp.MustCompile(`(?s)<h2.*?</h2>`)
	reHref      = regexp.MustCompile(`href="([^"]+)"`)
	reCite      = regexp.MustCompile(`(?s)<cite[^>]*>(.*?)</cite>`)
	rePara      = regexp.MustCompile(`(?s)<p[^>]*>(.*?)</p>`)
)

func (b *Bing) Search(ctx context.Context, query string, count int) ([]Result, error) {
	u := "https://www.bing.com/search?" + url.Values{
		"q": {query}, "count": {fmt.Sprint(countAt(count, b.Count))},
	}.Encode()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	httpc.Fingerprint(req.Header)
	res, err := b.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTransient, err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusTooManyRequests {
		return nil, ErrBlocked
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: status %d", ErrTransient, res.StatusCode)
	}
	body, _ := io.ReadAll(res.Body)
	if !reBingBlock.Match(body) {
		return nil, ErrBlocked
	}
	return parseBing(string(body)), nil
}

func parseBing(page string) []Result {
	out := []Result{}
	idx := regexp.MustCompile(`<li class="b_algo"`).FindAllStringIndex(page, -1)
	for i, start := range idx {
		end := len(page)
		if i+1 < len(idx) {
			end = idx[i+1][0]
		}
		block := page[start[0]:end]
		h2 := reH2.FindString(block)
		if h2 == "" {
			continue
		}
		title := html2text(h2)
		href := ""
		if m := reHref.FindStringSubmatch(h2); m != nil {
			href = html2text(m[1])
		}
		url := href
		if m := reCite.FindStringSubmatch(block); m != nil {
			url = strings.Split(html2text(m[1]), " › ")[0]
		}
		snippet := ""
		if m := rePara.FindStringSubmatch(block); m != nil {
			snippet = html2text(m[1])
		}
		if title != "" {
			out = append(out, Result{Title: title, URL: url, Snippet: snippet, Engine: "bing"})
		}
	}
	return out
}
