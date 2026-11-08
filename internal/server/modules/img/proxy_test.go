package img

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/FacileStudio/sonar/internal/server/imgsig"
)

func request(p *Proxy, raw, sig string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	target := "/img?" + url.Values{"u": {raw}, "s": {sig}}.Encode()
	p.serve(rec, httptest.NewRequest(http.MethodGet, target, nil))
	return rec
}

func TestProxyRefusesUnsignedAndInternalURLs(t *testing.T) {
	signer, _ := imgsig.New(nil)
	p := NewProxy(signer)
	if rec := request(p, "https://imgs.example/a.png", "forged"); rec.Code != http.StatusForbidden {
		t.Fatalf("unsigned: %d", rec.Code)
	}
	for _, internal := range []string{"http://127.0.0.1/a.png", "http://169.254.169.254/latest", "http://localhost/x"} {
		if rec := request(p, internal, signer.Sign(internal)); rec.Code != http.StatusForbidden {
			t.Fatalf("%s: %d", internal, rec.Code)
		}
	}
}

func response(body string) *http.Response {
	return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), ContentLength: -1}
}

func TestRelayServesOnlyRealImages(t *testing.T) {
	p := &Proxy{}
	png := "\x89PNG\r\n\x1a\n" + strings.Repeat("\x00", 64)
	rec := httptest.NewRecorder()
	p.relay(rec, response(png))
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "image/png" ||
		rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("png: %d %v", rec.Code, rec.Header())
	}
	for _, body := range []string{`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`,
		"<html><body>hi</body></html>"} {
		rec := httptest.NewRecorder()
		p.relay(rec, response(body))
		if rec.Code != http.StatusUnsupportedMediaType {
			t.Fatalf("%q served with %d", body[:10], rec.Code)
		}
	}
	big := &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("")), ContentLength: maxBytes + 1}
	rec = httptest.NewRecorder()
	p.relay(rec, big)
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("oversized: %d", rec.Code)
	}
}
