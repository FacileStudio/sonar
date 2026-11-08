package search

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/sonar/internal/server/imgsig"
	"github.com/FacileStudio/sonar/websearch"
	troncerrors "github.com/FacileStudio/tronc/errors"
)

type fakeEngines struct {
	calls int
	fail  error
	name  string
}

func (f *fakeEngines) Search(_ context.Context, q string, _ int, _ []string) ([]websearch.Result, error) {
	f.calls++
	return []websearch.Result{{Title: "hit for " + q, URL: "https://r.example/", Engine: f.name}}, f.fail
}

func (f *fakeEngines) Images(context.Context, string, int) ([]websearch.Image, error) {
	f.calls++
	return []websearch.Image{{Title: "img", URL: "https://p.example/", Thumbnail: "https://imgs.example/t.jpg"}}, f.fail
}

func (f *fakeEngines) News(context.Context, string, int) ([]websearch.News, error) {
	f.calls++
	return nil, f.fail
}

func (f *fakeEngines) Fetch(context.Context, string, websearch.FetchOptions) (*websearch.Page, error) {
	return &websearch.Page{}, f.fail
}

type fakePlans struct{ limit int64 }

func (p fakePlans) Daily(context.Context, int64) (int64, error) { return p.limit, nil }
func (p fakePlans) CheckAI(context.Context, int64) error        { return nil }
func (p fakePlans) Refusal(message string, cause error) error {
	return troncerrors.New("payment_required", message, cause).WithUpgradeURL("https://sso.example/spaces")
}

type fakeQuota struct{ used int64 }

func (q *fakeQuota) Take(_ context.Context, _ int64, limit int64) (bool, error) {
	if limit >= 0 && q.used >= limit {
		return false, nil
	}
	q.used++
	return true, nil
}
func (q *fakeQuota) Refund(ctx context.Context, _ int64) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	q.used--
	return nil
}
func (q *fakeQuota) Used(context.Context, int64) (int64, error) { return q.used, nil }
func (q *fakeQuota) ResetsAt() time.Time                        { return time.Time{} }

type fakeKeys map[string]string

func (k fakeKeys) Keys(context.Context, int64) (map[string]string, error) { return k, nil }

type harness struct {
	svc      *Service
	instance *fakeEngines
	own      *fakeEngines
	counter  *fakeQuota
}

func newHarness(limit int64, keys fakeKeys) harness {
	signer, _ := imgsig.New(nil)
	h := harness{instance: &fakeEngines{name: "instance"}, own: &fakeEngines{name: "own"}, counter: &fakeQuota{}}
	h.svc = &Service{
		Instance: h.instance, ForKeys: func(map[string]string) (Engines, error) { return h.own, nil },
		Plans: fakePlans{limit: limit}, Quota: h.counter, Keys: keys, Signer: signer, Logger: slog.Default(),
	}
	return h
}

func TestQuotaExhaustionAnswers402WithUpgradeURL(t *testing.T) {
	h := newHarness(2, nil)
	svc, instance := h.svc, h.instance
	for range 2 {
		if _, err := svc.Search(context.Background(), 1, Query{Q: "cats", Kind: KindWeb}); err != nil {
			t.Fatal(err)
		}
	}
	_, err := svc.Search(context.Background(), 1, Query{Q: "cats", Kind: KindWeb})
	var appErr *troncerrors.Error
	if troncerrors.Status(err) != 402 || !errors.As(err, &appErr) || appErr.UpgradeURL == "" {
		t.Fatalf("third search: %v", err)
	}
	if instance.calls != 2 {
		t.Fatalf("an exhausted quota must not reach the engines, calls = %d", instance.calls)
	}
}

func TestOwnKeysSkipTheQuota(t *testing.T) {
	h := newHarness(0, fakeKeys{"tavily": "k"})
	svc, instance, own, counter := h.svc, h.instance, h.own, h.counter
	response, err := svc.Search(context.Background(), 1, Query{Q: "cats", Kind: KindWeb})
	if err != nil || !response.OwnKeys || own.calls != 1 || instance.calls != 0 || counter.used != 0 {
		t.Fatalf("own-key search: %+v %v own=%d instance=%d used=%d", response, err, own.calls, instance.calls,
			counter.used)
	}
	if _, err := svc.Search(context.Background(), 1, Query{Q: "cats", Kind: KindImages}); troncerrors.Status(err) != 402 {
		t.Fatalf("images need a brave key; without one the quota applies: %v", err)
	}
}

func TestExternalBangRedirectsWithoutSpending(t *testing.T) {
	h := newHarness(5, nil)
	svc, instance, counter := h.svc, h.instance, h.counter
	response, err := svc.Search(context.Background(), 1, Query{Q: "!g private thing", Kind: KindWeb})
	if err != nil || !strings.HasPrefix(response.Redirect, "https://www.google.com/") || instance.calls != 0 ||
		counter.used != 0 {
		t.Fatalf("bang: %+v %v", response, err)
	}
	response, _ = svc.Search(context.Background(), 1, Query{Q: "cats !i", Kind: KindWeb})
	if response.Kind != KindImages || len(response.Images) != 1 || !strings.HasPrefix(response.Images[0].Thumbnail, "/img?") {
		t.Fatalf("internal bang must switch to images with proxied thumbnails: %+v", response)
	}
}

func TestFailedSearchIsRefunded(t *testing.T) {
	h := newHarness(5, nil)
	svc, instance, counter := h.svc, h.instance, h.counter
	instance.fail = errors.New("boom")
	if _, err := svc.Search(context.Background(), 1, Query{Q: "cats", Kind: KindWeb}); troncerrors.Status(err) != 503 {
		t.Fatalf("err = %v", err)
	}
	if counter.used != 0 {
		t.Fatalf("a failed search must be refunded, used = %d", counter.used)
	}
}

func TestAbortedSearchIsStillRefunded(t *testing.T) {
	h := newHarness(5, nil)
	h.instance.fail = errors.New("boom")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _ = h.svc.Search(ctx, 1, Query{Q: "cats", Kind: KindWeb})
	if h.counter.used != 0 {
		t.Fatalf("a search whose caller went away must be refunded, used = %d", h.counter.used)
	}
}
