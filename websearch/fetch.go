package websearch

import (
	"context"
	"errors"
	"time"
	"unicode/utf8"

	"github.com/FacileStudio/sonar/internal/engines"
	"github.com/FacileStudio/sonar/internal/netguard"
)

const (
	defaultFetchTimeout  = 30 * time.Second
	defaultFetchMaxBytes = 2 << 20
)

// Codes carried by a BlockedError, the same ones `sonar fetch --json` prints.
const (
	CodeBlockedURL        = netguard.CodeURL
	CodeBlockedInternal   = netguard.CodeInternal
	CodeBlockedUnresolved = netguard.CodeUnresolved
)

var newGuard = func() *netguard.Guard { return &netguard.Guard{} }

// ErrSelectorNeedsBrowser is returned by Fetch when Selector is set without
// Browser: the plain HTTP path cannot run CSS selectors.
var ErrSelectorNeedsBrowser = engines.ErrSelectorNeedsBrowser

// BlockedError is a refusal by the public-only guard. Code is one of the
// CodeBlocked* values. Reason never carries the resolved address.
type BlockedError struct {
	Code   string
	Reason string
}

// Error renders the code and the reason.
func (e *BlockedError) Error() string { return "websearch: " + e.Code + ": " + e.Reason }

// FetchOptions configures Fetch. The zero value is the safe default: plain
// HTTP, 30s timeout, 2 MiB cap and the public-only guard on.
//
// Browser renders the page in headless Chromium (go-rod). It needs a local
// Chromium; when none is found, rod downloads one. Without Browser, Fetch
// never starts a browser. Selector restricts extraction to a CSS selector
// and needs Browser.
//
// MaxBytes caps the HTML read on the HTTP path and the returned markdown.
//
// AllowInternal turns the public-only guard OFF. Set it only for trusted
// callers fetching trusted URLs: with it, a URL or any redirect can reach
// loopback, private networks and cloud metadata endpoints.
type FetchOptions struct {
	Selector      string
	Browser       bool
	Timeout       time.Duration
	MaxBytes      int64
	AllowInternal bool
}

// Page is one fetched page. URL is the URL asked for, FinalURL the one the
// page ended on after redirects, and Status the final HTTP status (0 when the
// browser did not report one). Engine is "http" or "rod". Truncated reports
// that the HTML or the markdown was cut to MaxBytes.
type Page struct {
	URL       string
	FinalURL  string
	Status    int
	Title     string
	Markdown  string
	Engine    string
	Truncated bool
}

// Fetch downloads one URL and returns it as markdown. Unless
// opts.AllowInternal is set, it applies the same guard as
// `sonar fetch --public-only`: the URL and every redirect hop are checked,
// every connection goes to an address that was checked and is re-checked at
// the dialer, and in browser mode every subrequest goes through the guard
// proxy. A refusal returns a *BlockedError and no content. A non-2xx status
// is not an error: check Page.Status.
func (c *Client) Fetch(ctx context.Context, rawURL string, opts FetchOptions) (*Page, error) {
	timeout := opts.Timeout
	if timeout <= 0 {
		timeout = defaultFetchTimeout
	}
	maxBytes := opts.MaxBytes
	if maxBytes <= 0 {
		maxBytes = defaultFetchMaxBytes
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	po := engines.PageOptions{Selector: opts.Selector, Browser: opts.Browser, Timeout: timeout, MaxBytes: maxBytes}
	if !opts.AllowInternal {
		po.Guard = newGuard()
	}
	res, err := engines.FetchPage(ctx, rawURL, po)
	if err != nil {
		return nil, publicError(err)
	}
	md, cut := truncate(res.Snippet, maxBytes)
	return &Page{
		URL: rawURL, FinalURL: res.FinalURL, Status: res.Status, Title: res.Title,
		Markdown: md, Engine: res.Engine, Truncated: cut || res.Cut,
	}, nil
}

func publicError(err error) error {
	if b, ok := netguard.AsBlocked(err); ok {
		return &BlockedError{Code: b.Code, Reason: b.Reason}
	}
	if errors.Is(err, engines.ErrSelectorNeedsBrowser) {
		return err
	}
	return errors.Join(errors.New("websearch: fetch failed"), err)
}

func truncate(s string, maxBytes int64) (string, bool) {
	if int64(len(s)) <= maxBytes {
		return s, false
	}
	cut := s[:maxBytes]
	for len(cut) > 0 && !utf8.ValidString(cut) {
		cut = cut[:len(cut)-1]
	}
	return cut, true
}
