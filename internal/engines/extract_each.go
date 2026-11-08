package engines

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/FacileStudio/sonar/internal/httpc"
	"github.com/FacileStudio/sonar/internal/netguard"
)

// Outcome is one URL's extraction: the page, or the error that stopped it.
type Outcome struct {
	URL    string
	Result Result
	Err    error
}

// ExtractEach extracts every URL and reports one Outcome per URL, in order.
// With opts.Guard set, a URL whose host is internal is refused before any
// request, and a page that reaches an internal address fails with a
// netguard.BlockedError instead of returning content.
func ExtractEach(ctx context.Context, urls []string, opts ExtractOptions) []Outcome {
	out := make([]Outcome, len(urls))
	var pending []int
	for i, u := range urls {
		out[i] = Outcome{URL: u, Err: precheck(ctx, u, opts.Guard)}
		if out[i].Err == nil {
			pending = append(pending, i)
		}
	}
	if len(pending) == 0 {
		return out
	}
	if res, ok := tryParallelExtract(ctx, pick(urls, pending), opts); ok {
		for j, i := range pending {
			out[i].Result = res[j]
		}
		return out
	}
	extractConcurrently(ctx, out, pending, opts)
	return out
}

func precheck(ctx context.Context, rawURL string, g *netguard.Guard) error {
	if g == nil {
		return nil
	}
	u, err := netguard.CheckURL(rawURL)
	if err != nil {
		return err
	}
	_, err = g.Resolve(ctx, u.Hostname())
	return err
}

func pick(urls []string, idx []int) []string {
	out := make([]string, len(idx))
	for j, i := range idx {
		out[j] = urls[i]
	}
	return out
}

func extractClient(g *netguard.Guard) *http.Client {
	if g != nil {
		return g.Client(15 * time.Second)
	}
	return httpc.Client(15 * time.Second)
}

func extractConcurrently(ctx context.Context, out []Outcome, pending []int, opts ExtractOptions) {
	client := extractClient(opts.Guard)
	sem := make(chan struct{}, 5)
	var wg sync.WaitGroup
	for _, i := range pending {
		wg.Go(func() {
			out[i].Result, out[i].Err = extractWorker(ctx, out[i].URL, opts, client, sem)
		})
	}
	wg.Wait()
}
