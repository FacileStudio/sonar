package netguard

import (
	"context"
	"io"
	"net"
	"net/http"
	"sync/atomic"
	"time"
)

// Proxy is a loopback forward proxy for the headless browser. Every request
// the browser makes (navigations, redirect hops, iframes, subresources,
// websockets) goes through it and reaches the network only through the
// guard, which connects to the address it checked.
type Proxy struct {
	guard     *Guard
	listener  net.Listener
	server    *http.Server
	transport *http.Transport
	blocked   atomic.Int64
}

var hopHeaders = []string{
	"Connection", "Proxy-Connection", "Keep-Alive", "Proxy-Authenticate",
	"Proxy-Authorization", "Te", "Trailer", "Transfer-Encoding", "Upgrade",
}

// StartProxy listens on a random loopback port and serves until Close.
func (g *Guard) StartProxy() (*Proxy, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	p := &Proxy{guard: g, listener: ln, transport: g.Transport()}
	p.server = &http.Server{Handler: p, ReadHeaderTimeout: 10 * time.Second}
	go func() { _ = p.server.Serve(ln) }()
	return p, nil
}

// Addr is the proxy's host:port.
func (p *Proxy) Addr() string { return p.listener.Addr().String() }

// Blocked counts the requests the guard refused so far for reaching an
// internal address or name. A host that merely does not resolve is not counted.
func (p *Proxy) Blocked() int64 { return p.blocked.Load() }

// Close stops the proxy and drops its idle upstream connections.
func (p *Proxy) Close() error {
	p.transport.CloseIdleConnections()
	return p.server.Close()
}

// ServeHTTP tunnels CONNECT requests and forwards plain HTTP ones.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		p.tunnel(w, r)
		return
	}
	if _, err := CheckURL(r.URL.String()); err != nil || r.URL.Scheme != "http" {
		p.refuse(w, err)
		return
	}
	out := r.Clone(r.Context())
	out.RequestURI = ""
	for _, h := range hopHeaders {
		out.Header.Del(h)
	}
	res, err := p.transport.RoundTrip(out)
	if err != nil {
		p.refuse(w, err)
		return
	}
	defer res.Body.Close()
	for _, h := range hopHeaders {
		res.Header.Del(h)
	}
	for k, v := range res.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(res.StatusCode)
	_, _ = io.Copy(w, res.Body)
}

func (p *Proxy) tunnel(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	upstream, err := p.guard.DialContext(ctx, "tcp", r.Host)
	if err != nil {
		p.refuse(w, err)
		return
	}
	hj, ok := w.(http.Hijacker)
	if !ok {
		_ = upstream.Close()
		http.Error(w, "tunnel unsupported", http.StatusInternalServerError)
		return
	}
	client, rw, err := hj.Hijack()
	if err != nil {
		_ = upstream.Close()
		return
	}
	_, _ = client.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
	go pipe(upstream, rw.Reader, client)
	pipe(client, upstream, upstream)
}

func pipe(dst io.WriteCloser, src io.Reader, closeAlso io.Closer) {
	_, _ = io.Copy(dst, src)
	_ = dst.Close()
	_ = closeAlso.Close()
}

func (p *Proxy) refuse(w http.ResponseWriter, err error) {
	b, ok := AsBlocked(err)
	if err == nil || (ok && b.Code != CodeUnresolved) {
		p.blocked.Add(1)
		http.Error(w, "blocked by sonar --public-only", http.StatusForbidden)
		return
	}
	http.Error(w, "upstream unreachable", http.StatusBadGateway)
}
