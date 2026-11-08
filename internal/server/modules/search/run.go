package search

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/FacileStudio/sonar/websearch"
	troncerrors "github.com/FacileStudio/tronc/errors"
)

const searchTimeout = 20 * time.Second

var errNoResults = errors.New("search: no results")

func (s *Service) run(ctx context.Context, engines Engines, q Query) (Response, error) {
	ctx, cancel := context.WithTimeout(ctx, searchTimeout)
	defer cancel()
	response := Response{Query: q.Q, Kind: q.Kind}
	var err error
	switch q.Kind {
	case KindImages:
		response.Images, err = s.images(ctx, engines, q)
	case KindNews:
		response.News, err = s.news(ctx, engines, q)
	default:
		response.Web, err = web(ctx, engines, q)
	}
	return response, engineError(err)
}

func web(ctx context.Context, engines Engines, q Query) ([]WebResult, error) {
	hits, err := engines.Search(ctx, q.Q, q.Count, q.Engines)
	out := make([]WebResult, 0, len(hits))
	for _, h := range hits {
		out = append(out, WebResult{Title: h.Title, URL: h.URL, Snippet: h.Snippet, Engine: h.Engine})
	}
	return out, err
}

func (s *Service) images(ctx context.Context, engines Engines, q Query) ([]ImageResult, error) {
	hits, err := engines.Images(ctx, q.Q, q.Count)
	out := make([]ImageResult, 0, len(hits))
	for _, h := range hits {
		thumb := s.Signer.Proxied(h.Thumbnail)
		if thumb == "" {
			continue
		}
		out = append(out, ImageResult{Title: h.Title, URL: h.URL, Thumbnail: thumb, Source: h.Source,
			Width: h.Width, Height: h.Height})
	}
	return out, err
}

func (s *Service) news(ctx context.Context, engines Engines, q Query) ([]NewsResult, error) {
	hits, err := engines.News(ctx, q.Q, q.Count)
	out := make([]NewsResult, 0, len(hits))
	for _, h := range hits {
		out = append(out, NewsResult{Title: h.Title, URL: h.URL, Snippet: h.Snippet, Source: h.Source,
			Age: h.Age, Published: h.Published, Thumbnail: s.Signer.Proxied(h.Thumbnail)})
	}
	return out, err
}

func engineError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, websearch.ErrNoEngines), errors.Is(err, websearch.ErrNoVertical):
		return troncerrors.Unavailable("no engine serves this tab right now")
	case strings.Contains(err.Error(), "no engine returned results"):
		return errNoResults
	case strings.Contains(err.Error(), "unknown or disabled engine"):
		return troncerrors.Invalid("that engine is not available for this search")
	}
	return troncerrors.Unavailable("the search engines did not answer; try again")
}
