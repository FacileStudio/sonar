// Package httpc builds guarded HTTP clients for scrape engines: a stable
// realistic browser fingerprint and a per-engine cookie jar.
package httpc

import (
	"net/http"
	"net/http/cookiejar"
	"time"
)

// uaDesktop is kept stable across requests. Randomly rotating User-Agents is
// itself a bot signal; a consistent, modern desktop UA reads as a real user.
const uaDesktop = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 " +
	"(KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36"

// Fingerprint applies a browser-like header set onto an outgoing request.
func Fingerprint(h http.Header) {
	h.Set("User-Agent", uaDesktop)
	h.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,"+
		"image/avif,image/webp,*/*;q=0.8")
	h.Set("Accept-Language", "en-US,en;q=0.9")
	h.Set("sec-ch-ua", `"Not/A)Brand";v="8", "Chromium";v="126", "Google Chrome";v="126"`)
	h.Set("sec-ch-ua-mobile", "?0")
	h.Set("Sec-Fetch-Dest", "document")
	h.Set("Sec-Fetch-Mode", "navigate")
	h.Set("Sec-Fetch-Site", "none")
	h.Set("Sec-Fetch-User", "?1")
	h.Set("Upgrade-Insecure-Requests", "1")
}

// Client returns an HTTP client that does not force a Connection: close and
// lets the response body be decoded. A per-process jar keeps cookies across
// queries so each engine sees one long-lived session instead of a fresh one.
func Client(timeout time.Duration) *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{Timeout: timeout, Jar: jar}
}
