package engines

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
	"time"

	"github.com/FacileStudio/sonar/internal/netguard"
)

const fakePublic = "93.184.215.14"

type fakeResolver map[string]string

func (f fakeResolver) LookupNetIP(_ context.Context, _, host string) ([]netip.Addr, error) {
	if a, ok := f[host]; ok {
		return []netip.Addr{netip.MustParseAddr(a)}, nil
	}
	return nil, errors.New("no such host")
}

func testGuard() *netguard.Guard {
	return &netguard.Guard{
		Resolver: fakeResolver{"public.test": fakePublic, "private.test": "10.128.0.1"},
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, network, strings.Replace(address, fakePublic, "127.0.0.1", 1))
		},
	}
}

const longText = "This public page carries enough words to pass the fast path length check. "

func page(body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = io.WriteString(w, "<html><head><title>t</title>"+body+"</html>")
	}
}

func trapServer(t *testing.T) string {
	t.Helper()
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	_, port, _ := net.SplitHostPort(srv.Listener.Addr().String())
	mux.HandleFunc("/secret", page("<body><main>TOPSECRET</main></body>"))
	mux.HandleFunc("/ok", page("<body><main>"+strings.Repeat(longText, 4)+"</main></body>"))
	mux.HandleFunc("/302", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://127.0.0.1:"+port+"/secret", http.StatusFound)
	})
	mux.HandleFunc("/js", page(`<body><main>wait</main><script>location="http://169.254.169.254/latest/";</script></body>`))
	mux.HandleFunc("/meta", page(`<meta http-equiv="refresh" content="0;url=http://private.test:`+port+`/secret"><body>x</body>`))
	mux.HandleFunc("/ws", page(`<body><main>`+longText+`</main><script>new WebSocket("ws://127.0.0.1:`+port+`/");</script></body>`))
	mux.HandleFunc("/frame", page(`<body><main>`+longText+`</main><iframe src="http://127.0.0.1:`+port+`/secret"></iframe></body>`))
	return "http://public.test:" + port
}

func extractGuarded(t *testing.T, url string, browser bool) Outcome {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	opts := ExtractOptions{Guard: testGuard(), ForceBrowser: browser, WaitCondition: "load", Timeout: 5 * time.Second}
	return ExtractEach(ctx, []string{url}, opts)[0]
}

func requireBlocked(t *testing.T, o Outcome) {
	t.Helper()
	if strings.Contains(o.Result.Snippet, "TOPSECRET") {
		t.Fatalf("%s leaked internal content", o.URL)
	}
	if _, ok := netguard.AsBlocked(o.Err); !ok {
		t.Fatalf("%s: want a guard refusal, got %v (%q)", o.URL, o.Err, o.Result.Snippet)
	}
}

func TestGuardedFetchRefusesInternalURLsUpFront(t *testing.T) {
	for _, u := range []string{"http://169.254.169.254/", "http://127.0.0.1/", "http://glouton-redis/", "http://private.test/"} {
		requireBlocked(t, extractGuarded(t, u, false))
	}
}

func TestGuardedHTTPPathRefusesRedirectToLoopback(t *testing.T) {
	requireBlocked(t, extractGuarded(t, trapServer(t)+"/302", false))
}

func TestGuardedFetchAllowsPublicPage(t *testing.T) {
	o := extractGuarded(t, trapServer(t)+"/ok", false)
	if o.Err != nil || !strings.Contains(o.Result.Snippet, "public page") {
		t.Fatalf("public page refused: %v", o.Err)
	}
}
