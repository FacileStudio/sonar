package engines

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSearxngSendsTheQueryInTheBodyNeverInTheURL(t *testing.T) {
	const query = "quiet heron marsh"
	var method, rawURL, body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method, rawURL = r.Method, r.URL.String()
		_ = r.ParseForm()
		body = r.PostForm.Get("q") + "|" + r.PostForm.Get("format")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, recordedShape)
	}))
	defer srv.Close()
	eng := &Searxng{Label: "searxng", BaseURL: srv.URL, Client: srv.Client()}
	got, err := eng.Search(context.Background(), query, 10)
	if err != nil || len(got) != 3 || got[0].Engine != "searxng" {
		t.Fatalf("got %d results, err %v", len(got), err)
	}
	if method != http.MethodPost || rawURL != "/search" || body != query+"|json" {
		t.Fatalf("method %q url %q body %q", method, rawURL, body)
	}
}

func TestSearxngErrorsNameNeitherTheQueryNorTheInstance(t *testing.T) {
	const query = "quiet heron marsh"
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "{nope")
	}))
	defer bad.Close()
	gone := httptest.NewServer(http.NotFoundHandler())
	gone.Close()
	for name, base := range map[string]string{"undecodable": bad.URL, "unreachable": gone.URL, "malformed": "http://bad host"} {
		eng := &Searxng{Label: "searxng", BaseURL: base, Client: http.DefaultClient}
		_, err := eng.Search(context.Background(), query, 10)
		if !errors.Is(err, ErrTransient) {
			t.Fatalf("%s: err %v, want a transient error", name, err)
		}
		for _, leak := range []string{"quiet", "heron", "127.0.0.1", "bad host", "/search"} {
			if strings.Contains(err.Error(), leak) {
				t.Fatalf("%s: error %q leaks %q", name, err, leak)
			}
		}
	}
}
