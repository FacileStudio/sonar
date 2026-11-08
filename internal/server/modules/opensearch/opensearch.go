package opensearch

import (
	"encoding/xml"
	"net/http"
	"strings"

	"github.com/FacileStudio/sonar/internal/server/bangs"
)

type urlTemplate struct {
	Type     string `xml:"type,attr"`
	Method   string `xml:"method,attr"`
	Template string `xml:"template,attr"`
}

type image struct {
	Width  int    `xml:"width,attr"`
	Height int    `xml:"height,attr"`
	Type   string `xml:"type,attr"`
	URL    string `xml:",chardata"`
}

type document struct {
	XMLName       xml.Name      `xml:"http://a9.com/-/spec/opensearch/1.1/ OpenSearchDescription"`
	ShortName     string        `xml:"ShortName"`
	Description   string        `xml:"Description"`
	InputEncoding string        `xml:"InputEncoding"`
	Image         image         `xml:"Image"`
	URLs          []urlTemplate `xml:"Url"`
	SearchForm    string        `xml:"http://www.mozilla.org/2006/browser/search/ SearchForm"`
}

// Description builds the OpenSearch document for the sonar deployed at base.
func Description(base string) ([]byte, error) {
	base = strings.TrimSuffix(base, "/")
	doc := document{
		ShortName:     "Sonar",
		Description:   "Facile Sonar: private web search, no history.",
		InputEncoding: "UTF-8",
		Image:         image{Width: 32, Height: 32, Type: "image/svg+xml", URL: base + "/favicon.svg"},
		URLs: []urlTemplate{
			{Type: "text/html", Method: "get", Template: base + "/search?q={searchTerms}"},
			{Type: "application/x-suggestions+json", Method: "get", Template: base + "/api/suggest?q={searchTerms}"},
			{Type: "application/opensearchdescription+xml", Method: "get", Template: base + "/opensearch.xml"},
		},
		SearchForm: base + "/",
	}
	body, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), body...), nil
}

func description(base string) http.HandlerFunc {
	body, err := Description(base)
	return func(w http.ResponseWriter, _ *http.Request) {
		if err != nil {
			http.Error(w, "unavailable", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/opensearchdescription+xml; charset=utf-8")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		_, _ = w.Write(body)
	}
}

func entry(app http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rest, bang, found := bangs.Parse(r.URL.Query().Get("q"))
		if found && bang.External() {
			w.Header().Set("Referrer-Policy", "no-referrer")
			w.Header().Set("Cache-Control", "no-store")
			http.Redirect(w, r, bang.Target(rest), http.StatusFound)
			return
		}
		app.ServeHTTP(w, r)
	}
}
