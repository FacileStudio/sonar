package engines

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const recordedShape = `{"query":"q","number_of_results":0,"results":[
{"url":"https://a.example/1","title":"One","content":"first","engine":"duckduckgo","score":2.0},
{"url":"https://a.example/2","title":"Two","content":"second","engine":"brave","score":1.0},
{"url":"","title":"no url","content":"dropped"}],"answers":[],"suggestions":[]}`

func jsonServer(hits *atomic.Int32, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprint(w, body)
	}))
}

func statusServer(hits *atomic.Int32, code int, kind, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", kind)
		w.WriteHeader(code)
		fmt.Fprint(w, body)
	}))
}

func TestPoolParsesRecordedShape(t *testing.T) {
	var hits atomic.Int32
	srv := jsonServer(&hits, recordedShape)
	defer srv.Close()
	pool := NewSearxPool([]string{srv.URL + "/"}, srv.Client(), PoolOptions{})
	got, err := pool.Search(context.Background(), "open source search", 10)
	if err != nil || len(got) != 2 {
		t.Fatalf("got %d results, err %v", len(got), err)
	}
	if got[0].Title != "One" || got[0].URL != "https://a.example/1" || got[0].Snippet != "first" || got[0].Engine != PoolName {
		t.Fatalf("unexpected first result %+v", got[0])
	}
	if pool.Name() != "searxpool" || pool.Size() != 1 {
		t.Fatalf("name %q size %d", pool.Name(), pool.Size())
	}
}

func TestPoolRefusalsFallThroughAndParkTheInstance(t *testing.T) {
	cases := map[string]func(*atomic.Int32) *httptest.Server{
		"forbidden": func(h *atomic.Int32) *httptest.Server { return statusServer(h, 403, "text/html", "no") },
		"limited":   func(h *atomic.Int32) *httptest.Server { return statusServer(h, 429, "text/html", "slow") },
		"captcha":   func(h *atomic.Int32) *httptest.Server { return statusServer(h, 200, "text/html", "<html>captcha") },
		"broken":    func(h *atomic.Int32) *httptest.Server { return statusServer(h, 200, "application/json", "{nope") },
		"down":      func(h *atomic.Int32) *httptest.Server { return statusServer(h, 502, "text/html", "bad") },
	}
	for name, build := range cases {
		var bad, good atomic.Int32
		badSrv, goodSrv := build(&bad), jsonServer(&good, recordedShape)
		pool := NewSearxPool([]string{badSrv.URL, goodSrv.URL}, http.DefaultClient, PoolOptions{})
		for range 4 {
			if got, err := pool.Search(context.Background(), "cats", 10); err != nil || len(got) != 2 {
				t.Fatalf("%s: got %d results, err %v", name, len(got), err)
			}
		}
		if bad.Load() != 1 || good.Load() != 4 {
			t.Fatalf("%s: bad instance hit %d times, good %d; want 1 and 4", name, bad.Load(), good.Load())
		}
		badSrv.Close()
		goodSrv.Close()
	}
}

func TestPoolRotatesAcrossHealthyInstances(t *testing.T) {
	var a, b atomic.Int32
	sa, sb := jsonServer(&a, recordedShape), jsonServer(&b, recordedShape)
	defer sa.Close()
	defer sb.Close()
	pool := NewSearxPool([]string{sa.URL, sb.URL}, http.DefaultClient, PoolOptions{})
	for range 4 {
		if _, err := pool.Search(context.Background(), "cats", 10); err != nil {
			t.Fatal(err)
		}
	}
	if a.Load() != 2 || b.Load() != 2 {
		t.Fatalf("hits %d and %d, want 2 and 2", a.Load(), b.Load())
	}
}

func TestPoolTriesAtMostMaxTries(t *testing.T) {
	var hits [3]atomic.Int32
	var bases []string
	for i := range hits {
		srv := statusServer(&hits[i], 403, "text/html", "no")
		defer srv.Close()
		bases = append(bases, srv.URL)
	}
	pool := NewSearxPool(bases, http.DefaultClient, PoolOptions{MaxTries: 2})
	_, err := pool.Search(context.Background(), "cats", 10)
	total := hits[0].Load() + hits[1].Load() + hits[2].Load()
	if !errors.Is(err, ErrBlocked) || total != 2 {
		t.Fatalf("err %v after %d attempts, want ErrBlocked after 2", err, total)
	}
	_, err = pool.Search(context.Background(), "cats", 10)
	if !errors.Is(err, ErrBlocked) || hits[0].Load()+hits[1].Load()+hits[2].Load() != 3 {
		t.Fatalf("second search err %v, want the third instance tried once and ErrBlocked", err)
	}
	if _, err = pool.Search(context.Background(), "cats", 10); !errors.Is(err, ErrTransient) {
		t.Fatalf("all parked: err %v, want ErrTransient", err)
	}
}

func TestPoolTimesOutSlowInstance(t *testing.T) {
	slow := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(300 * time.Millisecond):
		}
	}))
	defer slow.Close()
	var good atomic.Int32
	fast := jsonServer(&good, recordedShape)
	defer fast.Close()
	pool := NewSearxPool([]string{slow.URL, fast.URL}, http.DefaultClient, PoolOptions{AttemptTimeout: 50 * time.Millisecond})
	started := time.Now()
	got, err := pool.Search(context.Background(), "cats", 10)
	if err != nil || len(got) != 2 || time.Since(started) > time.Second {
		t.Fatalf("got %d results, err %v, after %v", len(got), err, time.Since(started))
	}
}

func TestPoolAnswersEmptyWithoutParking(t *testing.T) {
	var hits atomic.Int32
	srv := jsonServer(&hits, `{"results":[]}`)
	defer srv.Close()
	pool := NewSearxPool([]string{srv.URL}, http.DefaultClient, PoolOptions{})
	for range 2 {
		if _, err := pool.Search(context.Background(), "cats", 10); err == nil {
			t.Fatal("empty answer must not count as success")
		}
	}
	if hits.Load() != 2 {
		t.Fatalf("instance hit %d times, want 2", hits.Load())
	}
}

const needle = "zq-private-needle-query"

type spy struct{ url, form atomic.Value }

func (sp *spy) server(code int, kind string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sp.url.Store(r.URL.String() + r.Header.Get("Referer"))
		_ = r.ParseForm()
		sp.form.Store(r.PostForm.Get("q"))
		w.Header().Set("Content-Type", kind)
		w.WriteHeader(code)
		fmt.Fprint(w, needle)
	}))
}

func closedBase() string {
	dead := httptest.NewServer(http.NotFoundHandler())
	dead.Close()
	return dead.URL
}

func assertNoLeak(t *testing.T, base string) {
	t.Helper()
	pool := NewSearxPool([]string{base}, &http.Client{Timeout: time.Second}, PoolOptions{})
	_, err := pool.Search(context.Background(), needle, 5)
	if err == nil {
		t.Fatalf("%s: expected an error", base)
	}
	for _, secret := range []string{needle, base, "127.0.0.1"} {
		if strings.Contains(err.Error(), secret) {
			t.Fatalf("error leaks %q: %q", secret, err)
		}
	}
	if state := fmt.Sprintf("%+v %v", pool.members[0].breaker, pool.Name()); strings.Contains(state, needle) {
		t.Fatal("pool state carries the query")
	}
}

func TestPoolNeverShowsTheQueryOrTheRequestURL(t *testing.T) {
	var sp spy
	bases := []string{closedBase()}
	for _, c := range []struct {
		code int
		kind string
	}{{403, "text/html"}, {200, "text/html"}, {500, "text/plain"}, {200, "application/json"}} {
		srv := sp.server(c.code, c.kind)
		defer srv.Close()
		bases = append(bases, srv.URL)
	}
	for _, base := range bases {
		assertNoLeak(t, base)
	}
	if u, _ := sp.url.Load().(string); strings.Contains(u, needle) {
		t.Fatalf("query reached the request URL: %q", u)
	}
	if q, _ := sp.form.Load().(string); q != needle {
		t.Fatalf("the form body carries %q, want the query", q)
	}
}
