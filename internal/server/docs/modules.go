package docs

import (
	"net/http"

	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/sonar/internal/server/modules/keys"
	"github.com/FacileStudio/sonar/internal/server/modules/summary"
	"github.com/FacileStudio/tronc/apiref"
)

func authModule() apiref.Module {
	return apiref.Module{
		Name:        "auth",
		Description: "Sign-in through Registre, served by porte.",
		Routes: []apiref.Route{
			{Method: "GET", Path: porte.RouteLogin, Summary: "Start single sign-on", Auth: "public", Status: http.StatusFound},
			{Method: "GET", Path: porte.RouteCallback, Summary: "Finish single sign-on", Auth: "public", Status: http.StatusFound},
			{Method: "POST", Path: porte.RouteLogout, Summary: "Sign out", Auth: session},
			{Method: "GET", Path: "/auth/session", Summary: "The signed-in user, or null, and the enabled features",
				Auth: "public"},
		},
	}
}

func keysModule() apiref.Module {
	return apiref.Module{
		Name:        "keys",
		Description: "The caller's own engine API keys, sealed at rest. Searches on them skip the quota.",
		Routes: []apiref.Route{
			{Method: "GET", Path: "/keys", Summary: "List saved keys (hints only)", Auth: session,
				ResponseBody: keys.ListResponse{}},
			{Method: "PUT", Path: "/keys/{engine}", Summary: "Save a key", Auth: session + " plus X-Facile-CSRF on cookies",
				PathParams:  []apiref.Field{{Name: "engine", Type: "string", Description: "brave, exa, tavily, ..."}},
				RequestBody: keys.SetRequest{}, Status: http.StatusNoContent},
			{Method: "DELETE", Path: "/keys/{engine}", Summary: "Forget a key", Auth: session + " plus X-Facile-CSRF on cookies",
				PathParams: []apiref.Field{{Name: "engine", Type: "string", Description: "The engine name."}},
				Status:     http.StatusNoContent},
		},
	}
}

func summaryModule() apiref.Module {
	errs := []apiref.Error{authError,
		{Status: 402, Code: "payment_required", Description: "The plan does not include ai.access."},
		{Status: 404, Code: "not_found", Description: "Summaries are not enabled on this deployment."}}
	route := apiref.Route{Method: "POST", Path: "/summary", Summary: "Summarize results", Auth: session,
		RequestBody: summary.Request{}, ResponseBody: summary.Response{}, Errors: errs}
	return apiref.Module{
		Name:        "summary",
		Description: "AI summary of a result page, written by Bulle. Paid plans only (ai.access).",
		Routes:      []apiref.Route{route},
	}
}
