// Package bangs parses !bang shortcuts out of a query. An internal bang picks
// a tab or an engine inside sonar. An external bang sends the query to another
// site, which then sees it: the UI says so wherever bangs are listed.
package bangs

import (
	"net/url"
	"strings"
)

// Bang is one shortcut. Exactly one of Kind, Engine and URL is set: Kind
// switches the tab (web, images, news), Engine runs one engine, and URL is an
// external search page where %s stands for the escaped query.
type Bang struct {
	Trigger string `json:"trigger"`
	Name    string `json:"name"`
	Kind    string `json:"kind,omitempty"`
	Engine  string `json:"engine,omitempty"`
	URL     string `json:"url,omitempty"`
}

// External reports whether the bang leaves sonar.
func (b Bang) External() bool {
	return b.URL != ""
}

// Target is the external URL for query, or "" for an internal bang.
func (b Bang) Target(query string) string {
	if !b.External() {
		return ""
	}
	return strings.Replace(b.URL, "%s", url.QueryEscape(query), 1)
}

// List returns every bang, internal ones first.
func List() []Bang {
	return append(internal(), external()...)
}

// Parse finds the first known bang token in query, written "!g" anywhere
// between spaces, and returns the query without it. ok is false when the query
// carries no known bang, and the query comes back unchanged.
func Parse(query string) (rest string, bang Bang, ok bool) {
	fields := strings.Fields(query)
	for i, field := range fields {
		if len(field) < 2 || field[0] != '!' {
			continue
		}
		if found, known := lookup(strings.ToLower(field[1:])); known {
			rest := strings.Join(append(append([]string{}, fields[:i]...), fields[i+1:]...), " ")
			return rest, found, true
		}
	}
	return query, Bang{}, false
}

func lookup(trigger string) (Bang, bool) {
	for _, b := range List() {
		if b.Trigger == trigger {
			return b, true
		}
	}
	return Bang{}, false
}
