// Package imgsig signs the image URLs sonar hands to the browser, so the image
// proxy serves only images a search result named and never acts as an open
// proxy for arbitrary URLs.
package imgsig

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net/url"
)

// Signer holds the HMAC key. The key lives only in memory unless the operator
// pins one, so a restart invalidates links nobody can still be showing.
type Signer struct {
	key []byte
}

// New builds a Signer from key, or from 32 random bytes when key is shorter
// than 32 bytes.
func New(key []byte) (*Signer, error) {
	if len(key) >= 32 {
		return &Signer{key: key}, nil
	}
	random := make([]byte, 32)
	if _, err := rand.Read(random); err != nil {
		return nil, err
	}
	return &Signer{key: random}, nil
}

// Sign returns the base64url HMAC-SHA256 of raw.
func (s *Signer) Sign(raw string) string {
	mac := hmac.New(sha256.New, s.key)
	mac.Write([]byte(raw))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// Verify reports, in constant time, whether sig signs raw.
func (s *Signer) Verify(raw, sig string) bool {
	want, err := base64.RawURLEncoding.DecodeString(sig)
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, s.key)
	mac.Write([]byte(raw))
	return hmac.Equal(mac.Sum(nil), want)
}

// Proxied rewrites raw into the proxy path, or "" when raw is empty or not an
// https or http URL.
func (s *Signer) Proxied(raw string) string {
	parsed, err := url.Parse(raw)
	if raw == "" || err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return ""
	}
	return "/img?" + url.Values{"u": {raw}, "s": {s.Sign(raw)}}.Encode()
}
