package search

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/FacileStudio/sonar/internal/server/bangs"
	"github.com/FacileStudio/sonar/internal/server/imgsig"
	troncerrors "github.com/FacileStudio/tronc/errors"
)

const maxQuery = 400

// Service runs searches for signed-in users. A user who brought keys for the
// tab searches on them and spends no quota; everyone else spends one search
// of their daily quota on sonar's own keys.
type Service struct {
	Instance Engines
	Web      Chain
	Sources  Chain
	ForKeys  func(keys map[string]string) (Engines, error)
	Plans    Plans
	Quota    Quota
	Keys     Keys
	Signer   *imgsig.Signer
	Logger   *slog.Logger
}

// Query is one search request. Literal skips !bang parsing, for callers
// such as MCP clients that pass a model's text rather than a person's.
type Query struct {
	Q       string
	Kind    Kind
	Count   int
	Engines []string
	Literal bool
}

// Search runs q for userID. An external !bang returns its redirect and runs
// nothing; an internal one switches the tab or the engine.
func (s *Service) Search(ctx context.Context, userID int64, q Query) (Response, error) {
	if !q.Literal {
		rest, bang, found := bangs.Parse(q.Q)
		if found && bang.External() {
			return Response{Query: rest, Kind: q.Kind, Redirect: bang.Target(rest)}, nil
		}
		q = applyBang(q, rest, bang, found)
	}
	if err := validate(q); err != nil {
		return Response{}, err
	}
	engines, own, err := s.pick(ctx, userID, q.Kind)
	if err != nil {
		return Response{}, err
	}
	if !own {
		if err := s.spend(ctx, userID); err != nil {
			return Response{}, err
		}
	}
	response, err := s.run(ctx, engines, q, own)
	if err != nil && !own {
		s.refund(ctx, userID)
	}
	response.OwnKeys = own
	if errors.Is(err, errNoResults) {
		return response, nil
	}
	return response, err
}

func applyBang(q Query, rest string, bang bangs.Bang, found bool) Query {
	if !found {
		return q
	}
	q.Q = rest
	if bang.Kind != "" {
		q.Kind = Kind(bang.Kind)
	}
	if bang.Engine != "" {
		q.Kind, q.Engines = KindWeb, []string{bang.Engine}
	}
	return q
}

func validate(q Query) error {
	text := strings.TrimSpace(q.Q)
	if text == "" {
		return troncerrors.Invalid("q is required")
	}
	if utf8.RuneCountInString(text) > maxQuery || !utf8.ValidString(text) {
		return troncerrors.Invalid("q is longer than 400 characters")
	}
	switch q.Kind {
	case KindWeb, KindImages, KindNews:
		return nil
	}
	return troncerrors.Invalid("type must be web, images or news")
}

func (s *Service) pick(ctx context.Context, userID int64, kind Kind) (Engines, bool, error) {
	keys, err := s.Keys.Keys(ctx, userID)
	if err != nil {
		return nil, false, troncerrors.Internal("failed to read your engine keys", err)
	}
	if len(keys) == 0 || (kind != KindWeb && keys["brave"] == "") {
		return s.Instance, false, nil
	}
	engines, err := s.ForKeys(keys)
	if err != nil {
		return nil, false, troncerrors.Failed("one of your engine keys could not be used; check it in settings")
	}
	return engines, true, nil
}

func (s *Service) spend(ctx context.Context, userID int64) error {
	limit, err := s.Plans.Daily(ctx, userID)
	if err != nil {
		return troncerrors.Internal("failed to read your plan", err)
	}
	ok, err := s.Quota.Take(ctx, userID, limit)
	if err != nil {
		return troncerrors.Internal("failed to count the search", err)
	}
	if !ok {
		return s.Plans.Refusal("You used today's searches. Upgrade your plan, add your own engine key, "+
			"or come back after midnight UTC.", errQuota)
	}
	return nil
}

var errQuota = errors.New("search: daily quota exhausted")

func (s *Service) refund(ctx context.Context, userID int64) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	if err := s.Quota.Refund(ctx, userID); err != nil {
		s.Logger.Warn("search: failed to refund a search", slog.Int64("user", userID), slog.Any("error", err))
	}
}
