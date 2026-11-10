package mcp

import (
	"context"
	"net/url"
	"slices"
	"strings"

	"github.com/FacileStudio/sonar/internal/server/modules/search"
	troncerrors "github.com/FacileStudio/tronc/errors"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	defaultCount = 10
	maxCount     = 20
	noResults    = "No results. Try fewer or more common words, or another kind."
)

// search is sonar_search.
func (t *toolset) search(ctx context.Context, req *sdk.CallToolRequest, in searchInput) (*sdk.CallToolResult,
	searchOutput, error) {
	userID, err := userOf(req)
	if err != nil {
		return nil, searchOutput{}, err
	}
	query, err := in.query()
	if err != nil {
		return nil, searchOutput{}, err
	}
	if len(query.Engines) > 0 {
		if err := t.ownEngines(ctx, userID); err != nil {
			return nil, searchOutput{}, err
		}
	}
	response, err := t.searcher.Search(ctx, userID, query)
	if err != nil {
		return nil, searchOutput{}, mcpkit.Failure(err, "")
	}
	return nil, searchResult(response), nil
}

// query checks the arguments and turns them into a literal search.
func (in searchInput) query() (search.Query, error) {
	kind := search.Kind(strings.ToLower(strings.TrimSpace(in.Kind)))
	if kind == "" {
		kind = search.KindWeb
	}
	if kind != search.KindWeb && kind != search.KindImages && kind != search.KindNews {
		return search.Query{}, troncerrors.Invalid("kind must be web, images or news")
	}
	count := in.Count
	if count == 0 {
		count = defaultCount
	}
	if count < 1 || count > maxCount {
		return search.Query{}, troncerrors.Invalid("count must be between 1 and 20")
	}
	engines := slices.DeleteFunc(slices.Clone(in.Engines), func(e string) bool { return strings.TrimSpace(e) == "" })
	if len(engines) > 0 && kind != search.KindWeb {
		return search.Query{}, troncerrors.Invalid("engines applies to kind web only. Omit it for images and news")
	}
	return search.Query{Q: in.Query, Kind: kind, Count: count, Engines: engines, Literal: true}, nil
}

// ownEngines refuses an engine choice from a user who brought no keys:
// sonar's own keys always run the hosted engine policy.
func (t *toolset) ownEngines(ctx context.Context, userID int64) error {
	usage, err := t.searcher.Usage(ctx, userID)
	if err != nil {
		return mcpkit.Failure(err, "")
	}
	if len(usage.OwnKeys) == 0 {
		return troncerrors.Invalid("engines works only with your own engine keys. Omit engines, or ask the user " +
			"to add a key at https://sonar.facile.studio/settings/keys")
	}
	return nil
}

// searchResult keeps the fields a model needs from a search response.
func searchResult(r search.Response) searchOutput {
	out := searchOutput{Query: r.Query, Kind: string(r.Kind), OwnKeys: r.OwnKeys}
	for _, h := range r.Web {
		out.Results = append(out.Results, webHit{Title: h.Title, URL: h.URL, Snippet: h.Snippet})
	}
	for _, h := range r.Images {
		out.Images = append(out.Images, imageHit{Title: h.Title, URL: h.URL, ImageURL: unproxied(h.Thumbnail),
			Source: h.Source, Width: h.Width, Height: h.Height})
	}
	for _, h := range r.News {
		out.News = append(out.News, newsHit{Title: h.Title, URL: h.URL, Snippet: h.Snippet, Source: h.Source,
			Age: h.Age, Published: h.Published})
	}
	if len(out.Results)+len(out.Images)+len(out.News) == 0 {
		out.Note = noResults
	}
	return out
}

// unproxied is the image URL behind a signed /img path. The proxy needs a
// sonar session, which an MCP client does not hold.
func unproxied(proxied string) string {
	parsed, err := url.Parse(proxied)
	if err != nil {
		return ""
	}
	return parsed.Query().Get("u")
}
