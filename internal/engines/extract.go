package engines

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/FacileStudio/sonar/internal/httpc"
	"github.com/FacileStudio/sonar/internal/netguard"
)

// ExtractOptions configures web page extraction.
type ExtractOptions struct {
	ParallelKey   string
	WaitCondition string
	Selector      string
	Timeout       time.Duration
	ForceBrowser  bool
	Guard         *netguard.Guard
}

var reHTMLTitle = regexp.MustCompile(`(?i)<title\b[^>]*>(.*?)</title>`)

func extractTitle(body string) string {
	m := reHTMLTitle.FindStringSubmatch(body)
	if len(m) > 1 {
		return html2text(m[1])
	}
	return ""
}

func extractHTTP(ctx context.Context, rawURL string, client *http.Client) (Result, error) {
	res, err := httpPage(ctx, rawURL, client, 2*1024*1024)
	if err == nil && res.Status != http.StatusOK {
		return Result{}, fmt.Errorf("%w: status %d", ErrTransient, res.Status)
	}
	return res, err
}

func extractRod(ctx context.Context, rawURL string, opts ExtractOptions) (Result, error) {
	rodEngine := &Rod{
		Count:         1,
		WaitCondition: opts.WaitCondition,
		Selector:      opts.Selector,
		Timeout:       opts.Timeout,
		Guard:         opts.Guard,
	}
	results, err := rodEngine.Search(ctx, rawURL, 1)
	if err != nil {
		return Result{}, err
	}
	if len(results) == 0 {
		return Result{}, fmt.Errorf("%w: no content extracted", ErrTransient)
	}
	return results[0], nil
}

func extractOne(ctx context.Context, rawURL string, opts ExtractOptions, client *http.Client) (Result, error) {
	if opts.ForceBrowser || opts.Selector != "" {
		return extractRod(ctx, rawURL, opts)
	}
	res, err := extractHTTP(ctx, rawURL, client)
	if b, blocked := netguard.AsBlocked(err); blocked {
		return Result{}, b
	}
	if err == nil && len(strings.TrimSpace(res.Snippet)) >= 150 {
		return res, nil
	}
	return extractRod(ctx, rawURL, opts)
}

func tryParallelExtract(ctx context.Context, urls []string, opts ExtractOptions) ([]Result, bool) {
	if opts.ForceBrowser || opts.Selector != "" || opts.ParallelKey == "" || opts.Guard != nil {
		return nil, false
	}
	par := &Parallel{Key: opts.ParallelKey, Client: httpc.Client(20 * time.Second)}
	res, err := par.Extract(ctx, urls)
	if err == nil && len(res) == len(urls) {
		return res, true
	}
	return nil, false
}

func extractWorker(ctx context.Context, u string, opts ExtractOptions, client *http.Client, sem chan struct{}) (Result, error) {
	select {
	case sem <- struct{}{}:
		defer func() { <-sem }()
	case <-ctx.Done():
		return Result{}, ctx.Err()
	}
	return extractOne(ctx, u, opts, client)
}

// ExtractPages extracts markdown content from multiple URLs concurrently using
// the fastest available strategy. It returns the pages that succeeded, or the
// first error when none did.
func ExtractPages(ctx context.Context, urls []string, opts ExtractOptions) ([]Result, error) {
	if len(urls) == 0 {
		return nil, errors.New("no URLs to extract")
	}
	var out []Result
	var first error
	for _, o := range ExtractEach(ctx, urls, opts) {
		if o.Err == nil {
			out = append(out, o.Result)
		} else if first == nil {
			first = o.Err
		}
	}
	if len(out) == 0 {
		return nil, first
	}
	return out, nil
}
