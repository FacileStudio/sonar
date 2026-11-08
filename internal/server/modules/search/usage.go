package search

import (
	"context"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/FacileStudio/sonar/websearch"
	troncerrors "github.com/FacileStudio/tronc/errors"
)

const fetchMaxBytes = 512 << 10

// Usage reports the user's quota for today, the engines they brought keys for
// and whether their plan includes the AI features.
func (s *Service) Usage(ctx context.Context, userID int64) (Usage, error) {
	limit, err := s.Plans.Daily(ctx, userID)
	if err != nil {
		return Usage{}, troncerrors.Internal("failed to read your plan", err)
	}
	used, err := s.Quota.Used(ctx, userID)
	if err != nil {
		return Usage{}, troncerrors.Internal("failed to read your usage", err)
	}
	keys, err := s.Keys.Keys(ctx, userID)
	if err != nil {
		return Usage{}, troncerrors.Internal("failed to read your engine keys", err)
	}
	own := make([]string, 0, len(keys))
	for engine := range keys {
		own = append(own, engine)
	}
	sort.Strings(own)
	ai := s.Plans.CheckAI(ctx, userID) == nil
	return Usage{Used: used, Limit: limit, ResetsAt: s.Quota.ResetsAt(), OwnKeys: own, AI: ai}, nil
}

// Fetch downloads one public page as markdown. It spends one search of the
// quota, and the public-only guard refuses private and internal addresses.
func (s *Service) Fetch(ctx context.Context, userID int64, rawURL string) (FetchResponse, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" || len(rawURL) > 2048 {
		return FetchResponse{}, troncerrors.Invalid("url must be an http or https URL")
	}
	if err := s.spend(ctx, userID); err != nil {
		return FetchResponse{}, err
	}
	page, err := s.Instance.Fetch(ctx, parsed.String(), websearch.FetchOptions{Timeout: 20 * time.Second,
		MaxBytes: fetchMaxBytes})
	if err != nil {
		s.refund(ctx, userID)
		return FetchResponse{}, fetchError(err)
	}
	return FetchResponse{URL: page.URL, FinalURL: page.FinalURL, Title: page.Title, Markdown: page.Markdown,
		Truncated: page.Truncated}, nil
}

func fetchError(err error) error {
	if _, blocked := err.(*websearch.BlockedError); blocked {
		return troncerrors.Forbidden("that address is not on the public internet")
	}
	return troncerrors.Unavailable("the page could not be fetched")
}
