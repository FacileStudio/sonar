package summary

import (
	"context"
	stderrors "errors"
	"net/http"
	"time"

	"github.com/FacileStudio/tronc/bulle"
	troncerrors "github.com/FacileStudio/tronc/errors"
)

// Bulle writes summaries through Bulle's POST /api/ai/summarize as sonar's
// service: the key is sonar's service key in Bulle's BULLE_AI_CALLERS, and the
// acting person goes in X-Bulle-Acting-For. The client never follows a
// redirect.
type Bulle struct {
	client *bulle.Client
}

// NewBulle builds the Bulle writer, nil when either setting is missing, or
// an error when baseURL is not an http or https origin.
func NewBulle(baseURL, key string) (*Bulle, error) {
	if baseURL == "" || key == "" {
		return nil, nil
	}
	client, err := bulle.New(bulle.Config{
		URL:         baseURL,
		TokenSource: func(context.Context) (string, error) { return key, nil },
		HTTPClient:  &http.Client{Timeout: 30 * time.Second},
	})
	if err != nil {
		return nil, err
	}
	return &Bulle{client: client}, nil
}

// Summarize asks Bulle to summarize req for the person behind actingFor.
func (b *Bulle) Summarize(ctx context.Context, actingFor string, req Request) (Response, error) {
	var out Response
	err := b.client.Call(ctx, bulle.Request{
		Method: http.MethodPost, Segments: []string{"ai", "summarize"}, Body: req,
		Header: http.Header{bulle.ActingForHeader: {actingFor}},
	}, &out)
	return out, err
}

// failure turns a Writer error into the answer the person sees: Bulle's
// plan refusal with its upgrade link, its daily cap, a refusal that retrying
// will not fix, or a temporary outage.
func failure(err error) error {
	var plan *bulle.PlanRequired
	var refused *bulle.StatusError
	switch {
	case stderrors.As(err, &plan):
		return troncerrors.PaymentRequired(orDefault(plan.Message, "AI summaries need a paid plan.")).
			WithUpgradeURL(plan.UpgradeURL)
	case stderrors.As(err, &refused) && refused.Status == http.StatusTooManyRequests:
		return troncerrors.RateLimited(orDefault(refused.Message, "the daily AI limit is reached"))
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
