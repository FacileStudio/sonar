package docs

import (
	"github.com/FacileStudio/sonar/internal/server/modules/search"
	"github.com/FacileStudio/tronc/apiref"
)

func searchModule() apiref.Module {
	return apiref.Module{
		Name:        "search",
		Description: "Searches, page fetches and suggestions. Brave results are never cached.",
		Routes:      []apiref.Route{searchRoute(), fetchRoute(), suggestRoute(), usageRoute(), bangsRoute()},
	}
}

func searchParams() []apiref.Field {
	return []apiref.Field{
		{Name: "q", Type: "string", Description: "The query, 1 to 400 characters."},
		{Name: "type", Type: "string", Description: "web (default), images or news."},
		{Name: "count", Type: "integer", Description: "1 to 50, default 10."},
		{Name: "engines", Type: "string", Description: "Comma-separated engine names to run (web only)."},
	}
}

func searchErrors() []apiref.Error {
	return []apiref.Error{authError, quotaError,
		{Status: 400, Code: "invalid_argument", Description: "Empty or long query, or unknown type."},
		{Status: 429, Code: "rate_limited", Description: "Too many searches at once."},
		{Status: 503, Code: "unavailable", Description: "No engine answered; the search is refunded."}}
}

func searchRoute() apiref.Route {
	return apiref.Route{
		Method: "GET", Path: "/search", Summary: "Search the web, images or news",
		Description: "Runs on the caller's own engine keys when they cover the tab (no quota), otherwise " +
			"spends one search of the daily quota. A !bang picks a tab, an engine, or an external site, " +
			"in which case only redirect is set.",
		Auth: session, QueryParams: searchParams(), ResponseBody: search.Response{}, Errors: searchErrors(),
	}
}

func fetchRoute() apiref.Route {
	blocked := apiref.Error{Status: 403, Code: "permission_denied", Description: "The address is not public."}
	return apiref.Route{
		Method: "POST", Path: "/fetch", Summary: "Fetch a public page as markdown",
		Description: "Spends one search. Private, loopback and metadata addresses are refused at every hop.",
		Auth:        session, RequestBody: search.FetchRequest{}, ResponseBody: search.FetchResponse{},
		Errors: []apiref.Error{authError, quotaError, blocked},
	}
}

func suggestRoute() apiref.Route {
	return apiref.Route{
		Method: "GET", Path: "/suggest", Summary: "Complete a partial query",
		Description: "OpenSearch suggestions format: [query, [completions]]. Never counts against the quota.",
		Auth:        session,
		QueryParams: []apiref.Field{{Name: "q", Type: "string", Description: "The partial query."}},
	}
}

func usageRoute() apiref.Route {
	return apiref.Route{Method: "GET", Path: "/usage", Summary: "Today's quota", Auth: session,
		ResponseBody: search.Usage{}, Errors: []apiref.Error{authError}}
}

func bangsRoute() apiref.Route {
	return apiref.Route{Method: "GET", Path: "/bangs", Summary: "List the !bang shortcuts", Auth: "public",
		Description: "External bangs send the query to that site, which then sees it."}
}
