package docs

import (
	"github.com/FacileStudio/sonar/internal/server/modules/admin"
	"github.com/FacileStudio/tronc/apiref"
)

func adminModule() apiref.Module {
	route := apiref.Route{
		Method: "GET", Path: "/admin/engine-stats", Summary: "How each engine answered, per day",
		Description: "Counters per UTC day, kind of search (web, images, news, summary) and engine, for searches on " +
			"sonar's own keys: tried, answered (the search stopped on this engine's results), short (results, but too " +
			"few, and a later engine completed the answer), failed (error, block, captcha or nothing returned) and " +
			"timed_out. period adds the days up per kind, with each engine's share of the answered searches and the " +
			"split between free engines (no API key) and paid ones. No query, URL or user is stored or returned.",
		Auth: session + " of an administrator: a verified email listed in SONAR_ADMINS, or the Registre role " +
			admin.Role,
		QueryParams:  []apiref.Field{{Name: "days", Type: "integer", Description: "1 to 90 UTC days ending today, default 7."}},
		ResponseBody: admin.EngineStats{},
		Errors: []apiref.Error{authError,
			{Status: 403, Code: "permission_denied", Description: "The account's verified email is not on " +
				"SONAR_ADMINS and the session does not hold the admin role. An empty SONAR_ADMINS refuses everyone."},
			{Status: 400, Code: "invalid_argument", Description: "days is not between 1 and 90."}},
	}
	return apiref.Module{
		Name:        "admin",
		Description: "Read-only views for sonar's administrators.",
		Routes:      []apiref.Route{route},
	}
}
