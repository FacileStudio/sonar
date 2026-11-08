package keys

import (
	"errors"
	"net/http"

	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/sonar/internal/server/keyring"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/go-chi/chi/v5"
)

// ListResponse is the answer to GET /api/keys.
type ListResponse struct {
	Enabled bool            `json:"enabled"`
	Engines []string        `json:"engines"`
	Keys    []keyring.Saved `json:"keys"`
}

// SetRequest is the body of PUT /api/keys/{engine}.
type SetRequest struct {
	Key string `json:"key"`
}

type handler struct {
	ring *keyring.Ring
}

func (h *handler) list(w http.ResponseWriter, r *http.Request) {
	identity, _ := porte.From(r.Context())
	saved, err := h.ring.List(r.Context(), identity.UserID)
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to read your keys", err))
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	httpjson.WriteJSON(w, http.StatusOK, ListResponse{Enabled: h.ring.Enabled(), Engines: keyring.Engines, Keys: saved})
}

func (h *handler) set(w http.ResponseWriter, r *http.Request) {
	identity, _ := porte.From(r.Context())
	var body SetRequest
	if err := httpjson.DecodeJSONLimit(w, r, &body, 4<<10); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	if err := h.ring.Set(r.Context(), identity.UserID, chi.URLParam(r, "engine"), body.Key); err != nil {
		httpjson.WriteError(w, keyError(err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *handler) remove(w http.ResponseWriter, r *http.Request) {
	identity, _ := porte.From(r.Context())
	if err := h.ring.Delete(r.Context(), identity.UserID, chi.URLParam(r, "engine")); err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to delete the key", err))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func keyError(err error) error {
	switch {
	case errors.Is(err, keyring.ErrDisabled):
		return troncerrors.Unavailable("this server cannot store keys")
	case errors.Is(err, keyring.ErrUnknownEngine):
		return troncerrors.NotFound("that engine takes no key")
	case errors.Is(err, keyring.ErrBadKey):
		return troncerrors.Invalid("that does not look like an API key")
	}
	return troncerrors.Internal("failed to save the key", err)
}
