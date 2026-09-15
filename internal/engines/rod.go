package engines

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/stealth"
)

// Rod renders web pages in a headless browser via go-rod with stealth mode
// and extracts the main content text.
type Rod struct {
	Count         int
	WaitCondition string
}

// Name returns the engine identifier.
func (r *Rod) Name() string { return "rod" }

// Search navigates to the target URL and returns the extracted content as a Result.
func (r *Rod) Search(ctx context.Context, query string, count int) ([]Result, error) {
	if !strings.HasPrefix(query, "http://") && !strings.HasPrefix(query, "https://") {
		return nil, ErrBlocked
	}
	return r.doSearch(ctx, query)
}

func launchBrowser(ctx context.Context) (*rod.Browser, error) {
	u, err := launcher.New().Headless(true).Launch()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrTransient, err)
	}
	br := rod.New().ControlURL(u).Context(ctx)
	if err := br.Connect(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrTransient, err)
	}
	return br, nil
}

func (r *Rod) doSearch(ctx context.Context, query string) ([]Result, error) {
	br, err := launchBrowser(ctx)
	if err != nil {
		return nil, err
	}
	defer br.Close()

	page, err := stealth.Page(br)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrTransient, err)
	}
	page = page.Context(ctx)
	if err := page.Navigate(query); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrTransient, err)
	}
	if err := r.wait(page); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrTransient, err)
	}
	snippet, err := extractSnippet(page)
	if err != nil {
		return nil, err
	}
	res := Result{Title: query, Snippet: snippet, URL: query, Engine: "rod", Priority: 12}
	return []Result{res}, nil
}

func extractSnippet(page *rod.Page) (string, error) {
	raw, err := pageHTML(page)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrTransient, err)
	}
	snippet := html2text(raw)
	if snippet == "" {
		return "", fmt.Errorf("%w: no content extracted", ErrTransient)
	}
	return snippet, nil
}

func pageHTML(page *rod.Page) (string, error) {
	if ok, el, _ := page.Has("article"); ok && el != nil {
		if h, err := el.HTML(); err == nil && h != "" {
			return h, nil
		}
	}
	if ok, el, _ := page.Has("body"); ok && el != nil {
		if h, err := el.HTML(); err == nil && h != "" {
			return h, nil
		}
	}
	return page.HTML()
}

func (r *Rod) wait(page *rod.Page) error {
	cond := r.WaitCondition
	if cond == "" {
		cond = "networkidle"
	}
	switch cond {
	case "networkidle":
		page.Timeout(15*time.Second).WaitRequestIdle(500*time.Millisecond, nil, nil, nil)()
	case "domcontentloaded":
		if err := page.Timeout(15*time.Second).WaitDOMStable(100*time.Millisecond, 0.01); err != nil {
			return err
		}
	case "load":
		if err := page.WaitLoad(); err != nil {
			return err
		}
	case "none":
		return nil
	default:
		return fmt.Errorf("unknown wait condition %q", cond)
	}
	return nil
}
