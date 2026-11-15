package summary

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/FacileStudio/porte/exchange"
	"github.com/FacileStudio/tronc/bulle"
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
		w.WriteHeader(status)
		reply(t, w, body)
	}))
	defer srv.Close()
	writer, err := NewBulle(srv.URL, tokenOf("token"), nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = writer.Summarize(t.Context(), Request{Query: "q"})
	return err
}

func reply(t *testing.T, w io.Writer, body string) {
	t.Helper()
	if _, err := io.WriteString(w, body); err != nil {
		t.Error(err)
	}
}

func tokenOf(token string) bulle.TokenSource {
	return func(context.Context) (string, error) { return token, nil }
}

func TestWriterSendsTheExchangedTokenAndDecodesTheSummary(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/ai/summarize" || r.Header.Get("Authorization") != "Bearer exchanged" {
			t.Errorf("got %s with authorization %q", r.URL.Path, r.Header.Get("Authorization"))
		}
		if got := r.Header.Get("X-Bulle-Acting-For"); got != "" {
			t.Errorf("acting-for header is %q, want none", got)
		}
		var in bulle.SummarizeRequest
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil || in.Query != "q" || len(in.Sources) != 1 || in.Sources[0].Snippet != "s" {
			t.Errorf("request body %+v, err %v", in, err)
		}
		reply(t, w, `{"summary":"It works [1].","citations":[0],"model":"m"}`)
	}))
	defer srv.Close()
	writer, err := NewBulle(srv.URL, tokenOf("exchanged"), nil)
	if err != nil {
		t.Fatal(err)
	}
	got, err := writer.Summarize(t.Context(), Request{Query: "q", Sources: []Source{{Title: "t", URL: "https://example.com", Snippet: "s"}}})
	if err != nil || got.Summary != "It works [1]." || got.Model != "m" {
		t.Fatalf("got %+v, %v", got, err)
	}
}

func TestReconnectAsksToSignInAgain(t *testing.T) {
	writer, err := NewBulle("https://bulle.example", func(context.Context) (string, error) {
		return "", exchange.ErrReconnect
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = writer.Summarize(t.Context(), Request{Query: "q"})
	mapped, ok := failure(err).(*troncerrors.Error)
	if !ok || troncerrors.Status(mapped) >= http.StatusInternalServerError || !strings.Contains(mapped.Message, "Sign in again") {
		t.Fatalf("reconnect mapped to %+v", mapped)
	}
}

func TestRateLimitPassesBulleMessageThrough(t *testing.T) {
	body := `{"error":{"code":"rate_limited","message":"Monthly AI allowance used up; resets on 1 December."}}`
	mapped, ok := failure(summarizeAgainst(t, http.StatusTooManyRequests, body)).(*troncerrors.Error)
	if !ok || troncerrors.Status(mapped) != http.StatusTooManyRequests || mapped.Message != "Monthly AI allowance used up; resets on 1 December." {
		t.Fatalf("429 mapped to %+v", mapped)
	}
}
