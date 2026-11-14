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

func TestFitKeepsTheBodyUnderBulleCap(t *testing.T) {
	sources := make([]Source, maxSources)
	for i := range sources {
		sources[i] = Source{
			Title:   strings.Repeat("語", maxTitle),
			URL:     "https://example.com/" + strings.Repeat(string(rune('a'+i)), maxURL-20),
			Snippet: strings.Repeat("語", maxSnippet),
		}
	}
	got := fit(Request{Query: "q", Sources: sources})
	if n := encodedSize(got); n > maxBody {
		t.Fatalf("body is %d bytes, want at most %d", n, maxBody)
	}
	if len(got.Sources) == 0 || len(got.Sources) == maxSources {
		t.Fatalf("kept %d sources, want some dropped and some kept", len(got.Sources))
	}
	for i, source := range got.Sources {
		if source.URL != sources[i].URL {
			t.Fatalf("source %d is %q, want the original order kept", i, source.URL)
		}
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
		{429, "Rate limit exceeded\n", http.StatusTooManyRequests, ""},
		{400, `{"error":{"code":"invalid_argument","message":"bad source"}}`, http.StatusInternalServerError, ""},
		{503, ``, http.StatusServiceUnavailable, ""},
	}
	for _, c := range cases {
		mapped := failure(summarizeAgainst(t, c.status, c.body))
		if got := troncerrors.Status(mapped); got != c.want {
			t.Fatalf("Bulle %d mapped to %d, want %d", c.status, got, c.want)
		}
		e, ok := mapped.(*troncerrors.Error)
		if !ok || e.UpgradeURL != c.upgrade {
			t.Fatalf("Bulle %d: upgrade URL %+v, want %q", c.status, mapped, c.upgrade)
		}
		if c.status == http.StatusTooManyRequests && !strings.Contains(e.Message, "try again in a minute") {
			t.Fatalf("Bulle 429 says %q, want a retry hint", e.Message)
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
