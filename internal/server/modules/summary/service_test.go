package summary

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"

	troncerrors "github.com/FacileStudio/tronc/errors"
)

func TestCleanFitsBulleLimits(t *testing.T) {
	long := "https://example.com/?" + strings.Repeat("t", maxURL)
	got := clean([]Source{
		{Title: strings.Repeat("é", 400), URL: "https://example.com/a", Snippet: "s"},
		{Title: "tracked", URL: long, Snippet: "s"},
		{Title: "ftp", URL: "ftp://example.com", Snippet: "s"},
	})
	if len(got) != 1 || got[0].URL != "https://example.com/a" {
		t.Fatalf("clean kept %+v, want only the first source", got)
	}
	if n := utf8.RuneCountInString(got[0].Title); n != maxTitle || !utf8.ValidString(got[0].Title) {
		t.Fatalf("title is %d runes, want %d valid runes", n, maxTitle)
	}
}

func TestBulleFailures(t *testing.T) {
	cases := []struct {
		status  int
		body    string
		want    int
		upgrade string
	}{
		{402, `{"error":{"code":"payment_required","message":"Upgrade first.","upgrade_url":"https://sso.example/plans"}}`,
			http.StatusPaymentRequired, "https://sso.example/plans"},
		{429, `{"error":{"code":"rate_limited","message":"Daily cap."}}`, http.StatusTooManyRequests, ""},
		{400, `{"error":{"code":"invalid_argument","message":"bad source"}}`, http.StatusInternalServerError, ""},
		{503, ``, http.StatusServiceUnavailable, ""},
	}
	for _, c := range cases {
		mapped := failure(summarizeAgainst(t, c.status, c.body))
		if got := troncerrors.Status(mapped); got != c.want {
			t.Fatalf("Bulle %d mapped to %d, want %d", c.status, got, c.want)
		}
		if e, ok := mapped.(*troncerrors.Error); !ok || e.UpgradeURL != c.upgrade {
			t.Fatalf("Bulle %d: upgrade URL %+v, want %q", c.status, mapped, c.upgrade)
		}
	}
}

func summarizeAgainst(t *testing.T, status int, body string) error {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/ai/summarize" || r.Header.Get("X-Bulle-Acting-For") != "a@example.com" {
			t.Errorf("got %s acting for %q", r.URL.Path, r.Header.Get("X-Bulle-Acting-For"))
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()
	writer, err := NewBulle(srv.URL, "key")
	if err != nil {
		t.Fatal(err)
	}
	_, err = writer.Summarize(t.Context(), "a@example.com", Request{Query: "q"})
	return err
}
