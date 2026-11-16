package search

import (
	"context"
	"errors"
	"log/slog"
	"time"

	dispatch "github.com/FacileStudio/sonar/internal/search"
	"github.com/FacileStudio/sonar/internal/server/enginestats"
	"github.com/FacileStudio/sonar/websearch"
)

// KindSummary names, in the engine counters, the source search of an AI
// summary, which runs on its own chain.
const KindSummary = "summary"

const verticalEngine = "brave"

// Stats counts how each engine's turn ended. It receives the kind of search,
// engine names and outcomes, never the query.
type Stats interface {
	Record(ctx context.Context, kind string, attempts []enginestats.Attempt) error
}

// FreeEngine reports whether an engine runs without a provider API key:
// SearXNG instances, the SearXNG pool and the scrapers.
func FreeEngine(name string) bool {
	return !dispatch.NeedsKey(name)
}

func stageOutcome(ctx context.Context, err error, timedOut bool, added int, enough bool) enginestats.Outcome {
	switch {
	case err != nil && errors.Is(ctx.Err(), context.Canceled):
		return ""
	case err != nil && timedOut:
		return enginestats.TimedOut
	case err != nil || added == 0:
		return enginestats.Failed
	case enough:
		return enginestats.Answered
	}
	return enginestats.Short
}

func appendAttempt(tried []enginestats.Attempt, engine string, outcome enginestats.Outcome) []enginestats.Attempt {
	if outcome == "" {
		return tried
	}
	return append(tried, enginestats.Attempt{Engine: engine, Outcome: outcome})
}

func settle(tried []enginestats.Attempt, results int) []enginestats.Attempt {
	last := -1
	for i, attempt := range tried {
		if attempt.Outcome == enginestats.Answered {
			return tried
		}
		if attempt.Outcome == enginestats.Short {
			last = i
		}
	}
	if results > 0 && last >= 0 {
		tried[last].Outcome = enginestats.Answered
	}
	return tried
}

func (s *Service) record(ctx context.Context, kind string, tried []enginestats.Attempt) {
	if s.Stats == nil || len(tried) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 3*time.Second)
	go func() {
		defer cancel()
		if err := s.Stats.Record(ctx, kind, tried); err != nil {
			s.Logger.Warn("search: failed to count engine outcomes", slog.String("kind", kind), slog.Any("error", err))
		}
	}()
}

func (s *Service) recordVertical(ctx context.Context, kind Kind, results int, err error) {
	if kind == KindWeb || errors.Is(err, websearch.ErrNoVertical) {
		return
	}
	timedOut := errors.Is(ctx.Err(), context.DeadlineExceeded)
	outcome := stageOutcome(ctx, err, timedOut, results, true)
	s.record(ctx, string(kind), appendAttempt(nil, verticalEngine, outcome))
}
