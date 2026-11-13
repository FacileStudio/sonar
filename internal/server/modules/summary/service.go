package summary

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/FacileStudio/porte"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

const (
	maxSources = 10
	maxSnippet = 600
	maxTitle   = 300
	maxURL     = 2048
)

// Source is one result the summary may cite.
type Source struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// Ask is the body of POST /api/summary: only the query. sonar picks the
// sources itself, from the engines its summary chain allows.
type Ask struct {
	Query string `json:"query"`
}

// Request is what the Writer summarizes.
type Request struct {
	Query   string   `json:"query"`
	Sources []Source `json:"sources"`
}

// Response is the summary, citing Sources by their 1-based position.
type Response struct {
	Summary string   `json:"summary"`
	Model   string   `json:"model,omitempty"`
	Sources []Source `json:"sources,omitempty"`
}

// Access answers whether a user's plan includes ai.access.
type Access interface {
	CheckAI(ctx context.Context, userID int64) error
}

// Writer produces the summary for the person behind email.
type Writer interface {
	Summarize(ctx context.Context, actingFor string, req Request) (Response, error)
}

// Sources finds the results a summary may cite, spending the user's quota.
type Sources func(ctx context.Context, userID int64, query string, count int) ([]Source, error)

// Service gates and runs summaries. A nil Writer means the feature is off.
type Service struct {
	DB      *sql.DB
	Access  Access
	Writer  Writer
	Sources Sources
}

func (s *Service) handle(w http.ResponseWriter, r *http.Request) {
	if s.Writer == nil {
		httpjson.WriteError(w, troncerrors.NotFound("AI summaries are not available yet"))
		return
	}
	identity, _ := porte.From(r.Context())
	if err := s.Access.CheckAI(r.Context(), identity.UserID); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	var ask Ask
	if err := httpjson.DecodeJSONLimit(w, r, &ask, 4<<10); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	body, err := s.gather(r.Context(), identity.UserID, ask)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	response, err := s.run(r.Context(), identity.UserID, body)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	httpjson.WriteJSON(w, http.StatusOK, response)
}

func (s *Service) run(ctx context.Context, userID int64, body Request) (Response, error) {
	var email string
	if err := s.DB.QueryRowContext(ctx, `SELECT email FROM porte_users WHERE id = $1`, userID).Scan(&email); err != nil {
		return Response{}, troncerrors.Internal("failed to read your account", err)
	}
	response, err := s.Writer.Summarize(ctx, email, body)
	if err != nil {
		return Response{}, failure(err)
	}
	response.Sources = body.Sources
	return response, nil
}

func (s *Service) gather(ctx context.Context, userID int64, ask Ask) (Request, error) {
	query := strings.TrimSpace(ask.Query)
	if query == "" || utf8.RuneCountInString(query) > 400 {
		return Request{}, troncerrors.Invalid("query must be 1 to 400 characters")
	}
	found, err := s.Sources(ctx, userID, query, maxSources)
	if err != nil {
		return Request{}, err
	}
	sources := clean(found)
	if len(sources) == 0 {
		return Request{}, troncerrors.NotFound("no source was found to summarize")
	}
	return Request{Query: query, Sources: sources}, nil
}

func clean(found []Source) []Source {
	out := make([]Source, 0, min(len(found), maxSources))
	for _, source := range found {
		web := strings.HasPrefix(source.URL, "https://") || strings.HasPrefix(source.URL, "http://")
		if !web || len(source.URL) > maxURL {
			continue
		}
		source.Title = clip(source.Title, maxTitle)
		source.Snippet = clip(source.Snippet, maxSnippet)
		if out = append(out, source); len(out) == maxSources {
			break
		}
	}
	return out
}

func clip(text string, limit int) string {
	if utf8.RuneCountInString(text) > limit {
		return string([]rune(text)[:limit])
	}
	return text
}
