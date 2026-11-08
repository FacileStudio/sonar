package netguard

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func codeOf(err error) string {
	if b, ok := AsBlocked(err); ok {
		return b.Code
	}
	if err != nil {
		return "other"
	}
	return ""
}

func TestResolveRefusesAnyInternalAnswer(t *testing.T) {
	g := testGuard()
	cases := map[string]string{
		"public.test":  "",
		"mixed.test":   CodeInternal,
		"private.test": CodeInternal,
		"meta.test":    CodeInternal,
		"nowhere.test": CodeUnresolved,
		"glouton-db":   CodeURL,
	}
	for host, want := range cases {
		_, err := g.Resolve(context.Background(), host)
		if got := codeOf(err); got != want {
			t.Errorf("%s: got %q, want %q", host, got, want)
		}
	}
}

func TestControlChecksTheConnectedAddress(t *testing.T) {
	for _, a := range []string{"127.0.0.1:80", "[::1]:443", "10.0.0.1:6379", "169.254.169.254:80", "junk"} {
		if Control("tcp", a, nil) == nil {
			t.Errorf("%s should be refused", a)
		}
	}
	if err := Control("tcp4", fakePublic+":443", nil); err != nil {
		t.Errorf("public address refused: %v", err)
	}
}

func TestDefaultDialerRefusesLoopback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	defer srv.Close()
	_, err := guardedDialer.Dial("tcp", srv.Listener.Addr().String())
	if codeOf(err) != CodeInternal {
		t.Fatalf("dialer reached loopback: %v", err)
	}
}

func secretServer(t *testing.T) (*httptest.Server, string) {
	t.Helper()
	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	_, port, _ := net.SplitHostPort(srv.Listener.Addr().String())
	mux.HandleFunc("/secret", func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, "TOPSECRET") })
	mux.HandleFunc("/ok", func(w http.ResponseWriter, _ *http.Request) { _, _ = io.WriteString(w, "public page") })
	mux.HandleFunc("/to-loopback", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://127.0.0.1:"+port+"/secret", http.StatusFound)
	})
	mux.HandleFunc("/to-private-name", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://private.test:"+port+"/secret", http.StatusFound)
	})
	t.Cleanup(srv.Close)
	return srv, port
}

func TestClientRechecksRedirects(t *testing.T) {
	_, port := secretServer(t)
	c := testGuard().Client(5 * time.Second)
	for _, path := range []string{"/to-loopback", "/to-private-name"} {
		res, err := c.Get("http://public.test:" + port + path)
		if err == nil {
			body, _ := io.ReadAll(res.Body)
			_ = res.Body.Close()
			t.Fatalf("%s followed: %q", path, body)
		}
		if codeOf(err) == "" || codeOf(err) == "other" {
			t.Fatalf("%s: want a guard refusal, got %v", path, err)
		}
	}
	res, err := c.Get("http://public.test:" + port + "/ok")
	if err != nil {
		t.Fatalf("public page refused: %v", err)
	}
	_ = res.Body.Close()
}

func TestProxyRefusesInternalTargets(t *testing.T) {
	_, port := secretServer(t)
	p, err := testGuard().StartProxy()
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	proxyURL, _ := url.Parse("http://" + p.Addr())
	c := &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)}, Timeout: 5 * time.Second}
	for _, target := range []string{"http://127.0.0.1:" + port + "/secret", "http://private.test:" + port + "/secret"} {
		res, err := c.Get(target)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(res.Body)
		_ = res.Body.Close()
		if res.StatusCode != http.StatusForbidden || strings.Contains(string(body), "TOPSECRET") {
			t.Fatalf("%s: status %d body %q", target, res.StatusCode, body)
		}
	}
	res, err := c.Get("http://public.test:" + port + "/ok")
	if err != nil || res.StatusCode != http.StatusOK {
		t.Fatalf("public page through proxy: %v %v", res, err)
	}
	_ = res.Body.Close()
	if p.Blocked() != 2 {
		t.Fatalf("blocked = %d, want 2", p.Blocked())
	}
}

func connectStatus(t *testing.T, proxy, target string) string {
	t.Helper()
	conn, err := net.DialTimeout("tcp", proxy, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_, _ = io.WriteString(conn, "CONNECT "+target+" HTTP/1.1\r\nHost: "+target+"\r\n\r\n")
	buf := make([]byte, 64)
	n, _ := conn.Read(buf)
	return strings.SplitN(string(buf[:n]), "\r\n", 2)[0]
}

func TestProxyTunnelChecksTargets(t *testing.T) {
	_, port := secretServer(t)
	p, err := testGuard().StartProxy()
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	for _, target := range []string{"127.0.0.1:" + port, "private.test:" + port, "glouton-redis:6379"} {
		if got := connectStatus(t, p.Addr(), target); !strings.Contains(got, "403") {
			t.Errorf("CONNECT %s: %q", target, got)
		}
	}
	if got := connectStatus(t, p.Addr(), "public.test:"+port); !strings.Contains(got, "200") {
		t.Errorf("CONNECT public: %q", got)
	}
}
