// Package docs is the sonar API reference served at /docs, generated from a
// hand-written route inventory.
package docs

import (
	"github.com/FacileStudio/tronc/apiref"
)

const session = "session cookie or bearer token"

// Reference returns the reference configuration.
func Reference(version string) apiref.Config {
	return apiref.Config{
		Title:   "Sonar API",
		Version: version,
		Description: "Private multi-engine web search. Queries are never logged or stored; " +
			"each search on sonar's own engine keys counts once against a daily quota.",
		Servers:  []string{"/api"},
		Registry: apiref.Registry{Modules: []apiref.Module{authModule(), searchModule(), keysModule(), summaryModule()}},
	}
}

var quotaError = apiref.Error{Status: 402, Code: "payment_required",
	Description: "The daily quota is used up. upgrade_url names the page to change plan."}

var authError = apiref.Error{Status: 401, Code: "unauthenticated", Description: "No valid session."}
