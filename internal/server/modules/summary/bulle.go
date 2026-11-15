package summary

import (
	"context"
	stderrors "errors"
	"net/http"
	"time"

	"github.com/FacileStudio/porte/exchange"
	"github.com/FacileStudio/tronc/bulle"
	troncerrors "github.com/FacileStudio/tronc/errors"
)

// Bulle writes summaries through Bulle's POST /api/ai/summarize as the
// signed-in person: the token source exchanges the user on the request
// context for a Bulle token. The client never follows a redirect.
type Bulle struct {
	client *bulle.Client
}

// NewBulle builds the Bulle writer, nil when baseURL or source is missing,
// or an error when baseURL is not an http or https origin. forget drops the
// cached token before the one retry after a 401.
func NewBulle(baseURL string, source bulle.TokenSource, forget func(context.Context)) (*Bulle, error) {
	if baseURL == "" || source == nil {
		return nil, nil
	}
	client, err := bulle.New(bulle.Config{
		URL:         baseURL,
		TokenSource: source,
		Forget:      forget,
		HTTPClient:  &http.Client{Timeout: 30 * time.Second},
	})
	if err != nil {
		return nil, err
	}
	return &Bulle{client: client}, nil
}

// Summarize asks Bulle to summarize req for the person on ctx.
func (b *Bulle) Summarize(ctx context.Context, req Request) (Response, error) {
	sources := make([]bulle.SummarySource, len(req.Sources))
	for i, source := range req.Sources {
		sources[i] = bulle.SummarySource(source)
	}
	out, err := b.client.Summarize(ctx, bulle.SummarizeRequest{Query: req.Query, Sources: sources})
	return Response{Summary: out.Summary, Model: out.Model}, err
}

// failure turns a Writer error into the answer the person sees: a sign-in
// request when their session cannot be exchanged, Bulle's plan refusal with
// its upgrade link, its 429 (its own message when it has one, such as when
// the monthly allowance resets, else the burst limiter's plain text), a
// refusal that retrying will not fix, or a temporary outage.
func failure(err error) error {
	var plan *bulle.PlanRequired
	var refused *bulle.StatusError
	switch {
	case stderrors.Is(err, exchange.ErrReconnect):
		return troncerrors.Failed("Sign in again to use AI summaries.")
	case stderrors.As(err, &plan):
		return troncerrors.PaymentRequired(orDefault(plan.Message, "AI summaries need a paid plan.")).
			WithUpgradeURL(plan.UpgradeURL)
	case stderrors.As(err, &refused) && refused.Status == http.StatusTooManyRequests:
		return troncerrors.RateLimited(orDefault(refused.Message, "too many summaries; try again in a minute"))
	case stderrors.As(err, &refused):
		return troncerrors.Internal("the summary service refused the request", err)
	default:
		return troncerrors.Unavailable("the summary could not be written; try again")
	}
}

func orDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
