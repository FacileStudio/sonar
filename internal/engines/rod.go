package engines

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/stealth"
)

// Rod renders web pages in a headless browser via go-rod with stealth mode
// and extracts the main content as Markdown.
type Rod struct {
	Count         int
	WaitCondition string
	Selector      string
	Timeout       time.Duration
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

func newLauncher() *launcher.Launcher {
	return launcher.New().
		Headless(true).
		Leakless(true).
		Set("headless", "new").
		Set("disable-blink-features", "AutomationControlled").
		Set("exclude-switches", "enable-automation").
		Set("disable-infobars").
		Set("no-sandbox").
		Set("disable-setuid-sandbox").
		Set("disable-dev-shm-usage").
		Set("no-first-run").
		Set("no-default-browser-check").
		Set("window-size", "1920,1080").
		Set("start-maximized").
		Set("lang", "en-US,en")
}

func launchBrowser(ctx context.Context) (*launcher.Launcher, *rod.Browser, error) {
	l := newLauncher()
	u, err := l.Launch()
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %s", ErrTransient, err)
	}
	br := rod.New().ControlURL(u).Context(ctx)
	if err := br.Connect(); err != nil {
		l.Kill()
		return nil, nil, fmt.Errorf("%w: %s", ErrTransient, err)
	}
	return l, br, nil
}

func setupPage(br *rod.Browser, ctx context.Context) (*rod.Page, error) {
	page, err := stealth.Page(br)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrTransient, err)
	}
	page = page.Context(ctx)
	_ = page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		Width: 1920, Height: 1080, DeviceScaleFactor: 1, Mobile: false,
	})
	ua := "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36"
	_ = page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent: ua, AcceptLanguage: "en-US,en;q=0.9", Platform: "Linux x86_64",
		UserAgentMetadata: &proto.EmulationUserAgentMetadata{
			Brands: []*proto.EmulationUserAgentBrandVersion{
				{Brand: "Chromium", Version: "130"},
				{Brand: "Google Chrome", Version: "130"},
				{Brand: "Not?A_Brand", Version: "24"},
			},
			FullVersion: "130.0.6723.69", Platform: "Linux", Architecture: "x86", Mobile: false,
		},
	})
	return page, nil
}

func (r *Rod) doSearch(ctx context.Context, query string) ([]Result, error) {
	l, br, err := launchBrowser(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = br.Close()
		l.Kill()
	}()
	page, err := setupPage(br, ctx)
	if err != nil {
		return nil, err
	}
	if err := page.Navigate(query); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrTransient, err)
	}
	_ = r.wait(page)
	title, snippet, err := extractContent(page, r.Selector)
	if err != nil {
		return nil, err
	}
	if title == "" {
		title = query
	}
	return []Result{{Title: title, Snippet: snippet, URL: query, Engine: "rod", Priority: 12}}, nil
}

func (r *Rod) wait(page *rod.Page) error {
	cond := r.WaitCondition
	if cond == "" {
		cond = "networkidle"
	}
	if cond == "none" {
		return nil
	}
	if cond != "networkidle" && cond != "domcontentloaded" && cond != "load" {
		return fmt.Errorf("unknown wait condition %q", cond)
	}
	waitTimeout := r.Timeout
	if waitTimeout <= 0 {
		waitTimeout = 15 * time.Second
	}
	return rod.Try(func() {
		switch cond {
		case "networkidle":
			page.Timeout(waitTimeout).WaitRequestIdle(500*time.Millisecond, nil, nil, nil)()
		case "domcontentloaded":
			_ = page.Timeout(waitTimeout).WaitDOMStable(100*time.Millisecond, 0.01)
		case "load":
			_ = page.Timeout(waitTimeout).WaitLoad()
		}
	})
}
