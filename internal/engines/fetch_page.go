package engines

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/FacileStudio/sonar/internal/httpc"
	"github.com/FacileStudio/sonar/internal/netguard"
)

// ErrSelectorNeedsBrowser is returned by FetchPage when a selector is set
// without the browser: the plain HTTP path cannot run CSS selectors.
var ErrSelectorNeedsBrowser = errors.New("a selector needs the browser")

// PageOptions configures FetchPage. A nil Guard fetches without the
// public-only checks.
type PageOptions struct {
	Selector string
	Browser  bool
	Timeout  time.Duration
	MaxBytes int64
	Guard    *netguard.Guard
}

// FetchPage fetches one URL as markdown. Without Browser it uses plain HTTP
// only and never starts Chromium; any HTTP status is returned in the Result.
func FetchPage(ctx context.Context, rawURL string, opts PageOptions) (Result, error) {
	if opts.Selector != "" && !opts.Browser {
		return Result{}, ErrSelectorNeedsBrowser
	}
	if err := precheck(ctx, rawURL, opts.Guard); err != nil {
		return Result{}, err
	}
	if opts.Browser {
		return extractRod(ctx, rawURL, ExtractOptions{Selector: opts.Selector, Timeout: opts.Timeout, Guard: opts.Guard})
	}
	client := httpc.Client(opts.Timeout)
	if opts.Guard != nil {
		client = opts.Guard.Client(opts.Timeout)
	}
	return httpPage(ctx, rawURL, client, opts.MaxBytes)
}

func httpPage(ctx context.Context, rawURL string, client *http.Client, maxBytes int64) (Result, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return Result{}, fmt.Errorf("%w: %v", ErrTransient, err)
	}
	httpc.Fingerprint(req.Header)
	res, err := client.Do(req)
	if err != nil {
		return Result{}, fmt.Errorf("%w: %w", ErrTransient, err)
	}
	defer res.Body.Close()
	raw, cut, err := readCapped(res.Body, maxBytes)
	if err != nil {
		return Result{}, err
	}
	body := string(raw)
	md, err := HTMLToMarkdown(body)
	if err != nil {
		return Result{}, fmt.Errorf("%w: markdown: %v", ErrTransient, err)
	}
	title := extractTitle(body)
	if title == "" {
		title = rawURL
	}
	return Result{
		Title: title, URL: rawURL, Snippet: md, Engine: "http",
		FinalURL: res.Request.URL.String(), Status: res.StatusCode, Cut: cut,
	}, nil
}

func readCapped(r io.Reader, maxBytes int64) ([]byte, bool, error) {
	raw, err := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if err != nil {
		return nil, false, fmt.Errorf("%w: read: %v", ErrTransient, err)
	}
	if int64(len(raw)) > maxBytes {
		return raw[:maxBytes], true, nil
	}
	return raw, false, nil
}
