package websearch

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"

	"github.com/go-rod/rod/lib/launcher"

	"github.com/FacileStudio/sonar/internal/netguard"
)

const fakePublic = "93.184.215.14"

type fakeResolver struct{}

func (fakeResolver) LookupNetIP(_ context.Context, _, host string) ([]netip.Addr, error) {
	if host == "public.test" {
		return []netip.Addr{netip.MustParseAddr(fakePublic)}, nil
	}
	return nil, errors.New("no such host")
}

func useTestGuard(t *testing.T) {
	t.Helper()
	prev := newGuard
	newGuard = func() *netguard.Guard {
		return &netguard.Guard{Resolver: fakeResolver{}, Dial: func(ctx context.Context, network, addr string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, network, strings.Replace(addr, fakePublic, "127.0.0.1", 1))
		}}
	}
	t.Cleanup(func() { newGuard = prev })
}

func html(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = io.WriteString(w, "<html><head><title>T</title></head><body><main>"+body+"</main></body></html>")
	}
}

func fetchServer(t *testing.T) (public, loopback string) {
	t.Helper()
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	_, port, _ := net.SplitHostPort(srv.Listener.Addr().String())
	mux.HandleFunc("/ok", html("public page "+strings.Repeat("words ", 50)))
	mux.HandleFunc("/secret", html("TOPSECRET"))
	mux.HandleFunc("/hop", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/ok", http.StatusFound) })
	mux.HandleFunc("/302", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://127.0.0.1:"+port+"/secret", http.StatusFound)
	})
	mux.HandleFunc("/js", html(`wait<script>location="http://169.254.169.254/";</script>`))
	mux.HandleFunc("/gone", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNotFound) })
	return "http://public.test:" + port, "http://127.0.0.1:" + port
}

func blockedCode(err error) string {
	var b *BlockedError
	if errors.As(err, &b) {
		return b.Code
	}
	return ""
}

func TestFetchIsPublicOnlyByDefault(t *testing.T) {
	_, loopback := fetchServer(t)
	c, _ := New(Config{})
	page, err := c.Fetch(context.Background(), loopback+"/secret", FetchOptions{})
	if page != nil || blockedCode(err) != CodeBlockedInternal {
		t.Fatalf("loopback fetched by default: %+v %v", page, err)
	}
	for _, u := range []string{"http://169.254.169.254/", "http://glouton-redis/", "file:///etc/passwd"} {
		if _, err := c.Fetch(context.Background(), u, FetchOptions{}); blockedCode(err) == "" {
			t.Errorf("%s: want a refusal, got %v", u, err)
		}
	}
}

func TestFetchAllowInternalOptsOut(t *testing.T) {
	_, loopback := fetchServer(t)
	c, _ := New(Config{})
	page, err := c.Fetch(context.Background(), loopback+"/secret", FetchOptions{AllowInternal: true})
	if err != nil || !strings.Contains(page.Markdown, "TOPSECRET") {
		t.Fatalf("trusted fetch failed: %v", err)
	}
}

func TestFetchHTTPPathGuardsRedirects(t *testing.T) {
	useTestGuard(t)
	public, _ := fetchServer(t)
	c, _ := New(Config{})
	if _, err := c.Fetch(context.Background(), public+"/302", FetchOptions{}); blockedCode(err) != CodeBlockedInternal {
		t.Fatalf("redirect to loopback not refused: %v", err)
	}
	page, err := c.Fetch(context.Background(), public+"/hop", FetchOptions{})
	if err != nil || page.Status != 200 || page.FinalURL != public+"/ok" || page.Engine != "http" || page.Title != "T" {
		t.Fatalf("public redirect: %+v %v", page, err)
	}
	page, err = c.Fetch(context.Background(), public+"/gone", FetchOptions{})
	if err != nil || page.Status != http.StatusNotFound {
		t.Fatalf("404 page: %+v %v", page, err)
	}
}

func TestFetchOptions(t *testing.T) {
	useTestGuard(t)
	public, _ := fetchServer(t)
	c, _ := New(Config{})
	if _, err := c.Fetch(context.Background(), public+"/ok", FetchOptions{Selector: "main"}); !errors.Is(err, ErrSelectorNeedsBrowser) {
		t.Fatalf("selector without browser: %v", err)
	}
	page, err := c.Fetch(context.Background(), public+"/ok", FetchOptions{MaxBytes: 40})
	if err != nil || !page.Truncated || len(page.Markdown) > 40 {
		t.Fatalf("max bytes: %+v %v", page, err)
	}
}

func TestFetchBrowserPathGuardsSubrequests(t *testing.T) {
	if _, ok := launcher.LookPath(); !ok && launcher.NewBrowser().Validate() != nil {
		t.Skip("no local chromium")
	}
	useTestGuard(t)
	public, _ := fetchServer(t)
	c, _ := New(Config{})
	for _, path := range []string{"/302", "/js"} {
		page, err := c.Fetch(context.Background(), public+path, FetchOptions{Browser: true})
		if page != nil || blockedCode(err) != CodeBlockedInternal {
			t.Fatalf("%s via browser: %+v %v", path, page, err)
		}
	}
	page, err := c.Fetch(context.Background(), public+"/hop", FetchOptions{Browser: true, Selector: "main"})
	if err != nil || page.Engine != "rod" || page.Status != 200 || page.FinalURL != public+"/ok" {
		t.Fatalf("public page via browser: %+v %v", page, err)
	}
}
