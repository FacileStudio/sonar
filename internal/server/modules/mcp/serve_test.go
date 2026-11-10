package mcp

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"testing"
)

// TestServe serves the endpoint on MCP_SERVE until interrupted, for the
// official conformance suite and MCP Inspector. Requests without a token get
// a valid one, because the conformance suite sends none.
func TestServe(t *testing.T) {
	addr := os.Getenv("MCP_SERVE")
	if addr == "" {
		t.Skip("set MCP_SERVE=127.0.0.1:<port> to serve the endpoint")
	}
	f := newFixture(t, -1, nil)
	f.instance.markdown = "# Example\n\nA page."
	target, _ := url.Parse(f.resource)
	proxy := httputil.NewSingleHostReverseProxy(&url.URL{Scheme: target.Scheme, Host: target.Host})
	token := f.token(t)
	server := &http.Server{Addr: addr, Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		r.Header.Del("Origin")
		proxy.ServeHTTP(w, r)
	})}
	go func() { _ = server.ListenAndServe() }()
	t.Logf("serving http://%s/mcp", addr)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	_ = server.Close()
}
