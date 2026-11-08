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
)

// Source is one result the summary may cite.
type Source struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// Request is the body of POST /api/summary.
type Request struct {
	Query   string   `json:"query"`
	Sources []Source `json:"sources"`
}

// Response is the summary, citing sources by their 1-based position.
type Response struct {
	Summary string `json:"summary"`
	Model   string `json:"model,omitempty"`
}

// Access answers whether a user's plan includes ai.access.
type Access interface {
	CheckAI(ctx context.Context, userID int64) error
}

// Writer produces the summary for the person behind email.
type Writer interface {
	Summarize(ctx context.Context, actingFor string, req Request) (Response, error)
}

// Service gates and runs summaries. A nil Writer means the feature is off.
type Service struct {
	DB     *sql.DB
	Access Access
	Writer Writer
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
	var body Request
	if err := httpjson.DecodeJSONLimit(w, r, &body, 32<<10); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	if err := clean(&body); err != nil {
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
		return Response{}, troncerrors.Unavailable("the summary could not be written; try again")
	}
	return response, nil
}

func clean(body *Request) error {
	body.Query = strings.TrimSpace(body.Query)
	if body.Query == "" || utf8.RuneCountInString(body.Query) > 400 {
		return troncerrors.Invalid("query must be 1 to 400 characters")
	}
	if len(body.Sources) == 0 || len(body.Sources) > maxSources {
		return troncerrors.Invalid("send between 1 and 10 sources")
	}
	for i := range body.Sources {
		if !strings.HasPrefix(body.Sources[i].URL, "https://") && !strings.HasPrefix(body.Sources[i].URL, "http://") {
			return troncerrors.Invalid("every source needs an http or https url")
		}
		if utf8.RuneCountInString(body.Sources[i].Snippet) > maxSnippet {
			body.Sources[i].Snippet = string([]rune(body.Sources[i].Snippet)[:maxSnippet])
		}
	}
	return nil
}
