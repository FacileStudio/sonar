package search

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/sonar/internal/server/bangs"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

type handler struct {
	service *Service
	suggest *Suggester
}

func userID(r *http.Request) (int64, bool) {
	identity, ok := porte.From(r.Context())
	return identity.UserID, ok && identity.UserID > 0
}

func (h *handler) search(w http.ResponseWriter, r *http.Request) {
	id, ok := userID(r)
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("sign in to search"))
		return
	}
	query, err := parseQuery(r)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	response, err := h.service.Search(r.Context(), id, query)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	httpjson.WriteJSON(w, http.StatusOK, response)
}

func parseQuery(r *http.Request) (Query, error) {
	values := r.URL.Query()
	kind := Kind(values.Get("type"))
	if kind == "" {
		kind = KindWeb
	}
	count := 10
	if raw := values.Get("count"); raw != "" {
		n, err := strconv.Atoi(raw)
		if err != nil || n < 1 || n > 50 {
			return Query{}, troncerrors.Invalid("count must be between 1 and 50")
		}
		count = n
	}
	var engines []string
	if raw := strings.TrimSpace(values.Get("engines")); raw != "" {
		engines = strings.Split(raw, ",")
	}
	return Query{Q: values.Get("q"), Kind: kind, Count: count, Engines: engines}, nil
}

func (h *handler) fetch(w http.ResponseWriter, r *http.Request) {
	id, ok := userID(r)
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("sign in to fetch"))
		return
	}
	var body FetchRequest
	if err := httpjson.DecodeJSONLimit(w, r, &body, 8<<10); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	page, err := h.service.Fetch(r.Context(), id, body.URL)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	httpjson.WriteJSON(w, http.StatusOK, page)
}

func (h *handler) usage(w http.ResponseWriter, r *http.Request) {
	id, ok := userID(r)
	if !ok {
		httpjson.WriteError(w, troncerrors.Unauthorized("sign in"))
		return
	}
	usage, err := h.service.Usage(r.Context(), id)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	httpjson.WriteJSON(w, http.StatusOK, usage)
}

func (h *handler) bangs(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Cache-Control", "public, max-age=3600")
	httpjson.WriteJSON(w, http.StatusOK, map[string]any{"bangs": bangs.List()})
}
