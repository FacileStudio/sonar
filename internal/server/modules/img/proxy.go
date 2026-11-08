package img

import (
	"bufio"
	"io"
	"net/http"
	"slices"
	"strings"

	"github.com/FacileStudio/sonar/internal/netguard"
)

var allowedTypes = []string{"image/avif", "image/gif", "image/jpeg", "image/png", "image/webp"}

func (p *Proxy) serve(w http.ResponseWriter, r *http.Request) {
	raw, sig := r.URL.Query().Get("u"), r.URL.Query().Get("s")
	if raw == "" || !p.Signer.Verify(raw, sig) {
		http.Error(w, "unsigned image", http.StatusForbidden)
		return
	}
	if _, err := netguard.CheckURL(raw); err != nil {
		http.Error(w, "image host not allowed", http.StatusForbidden)
		return
	}
	res, err := p.fetch(r, raw)
	if err != nil {
		http.Error(w, "image unavailable", http.StatusBadGateway)
		return
	}
	defer func() { _ = res.Body.Close() }()
	p.relay(w, res)
}

func (p *Proxy) fetch(r *http.Request, raw string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, raw, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "image/avif,image/webp,image/png,image/jpeg,image/gif;q=0.8")
	req.Header.Set("User-Agent", "sonar-image-proxy/1 (+https://sonar.facile.studio)")
	return p.Client.Do(req)
}

func (p *Proxy) relay(w http.ResponseWriter, res *http.Response) {
	if res.StatusCode != http.StatusOK || res.ContentLength > maxBytes {
		http.Error(w, "image unavailable", http.StatusBadGateway)
		return
	}
	body := bufio.NewReaderSize(io.LimitReader(res.Body, maxBytes+1), 512)
	head, _ := body.Peek(512)
	kind := sniff(head)
	if kind == "" {
		http.Error(w, "not an image", http.StatusUnsupportedMediaType)
		return
	}
	header := w.Header()
	header.Set("Content-Type", kind)
	header.Set("Cache-Control", "private, max-age=86400")
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set("Content-Security-Policy", "default-src 'none'; sandbox")
	header.Set("Referrer-Policy", "no-referrer")
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, io.LimitReader(body, maxBytes))
}

func sniff(head []byte) string {
	kind := strings.SplitN(http.DetectContentType(head), ";", 2)[0]
	if slices.Contains(allowedTypes, kind) {
		return kind
	}
	if len(head) >= 12 && string(head[4:12]) == "ftypavif" {
		return "image/avif"
	}
	return ""
}
