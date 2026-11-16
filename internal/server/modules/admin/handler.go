package admin

import (
	"net/http"
	"strconv"

	"github.com/FacileStudio/sonar/internal/server/enginestats"
	"github.com/FacileStudio/sonar/internal/server/modules/search"
	troncerrors "github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
)

const (
	defaultDays = 7
	maxDays     = 90
)

// EngineStats is the answer to GET /api/admin/engine-stats: the daily
// counters of each engine on sonar's own keys, then the period added up per
// kind of search. It carries no query, no URL and no user.
type EngineStats struct {
	Days   int                     `json:"days"`
	From   string                  `json:"from"`
	To     string                  `json:"to"`
	Rows   []enginestats.Row       `json:"rows"`
	Period []enginestats.KindTotal `json:"period"`
}

type handler struct {
	stats *enginestats.Store
}

func (h *handler) engineStats(w http.ResponseWriter, r *http.Request) {
	days, err := parseDays(r.URL.Query().Get("days"))
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	rows, err := h.stats.Read(r.Context(), days)
	if err != nil {
		httpjson.WriteError(w, troncerrors.Internal("failed to read the engine counters", err))
		return
	}
	for i := range rows {
		rows[i].Free = search.FreeEngine(rows[i].Engine)
	}
	w.Header().Set("Cache-Control", "no-store")
	httpjson.WriteJSON(w, http.StatusOK, EngineStats{
		Days: days, From: h.stats.Since(days).Format("2006-01-02"), To: h.stats.Since(1).Format("2006-01-02"),
		Rows: rows, Period: enginestats.Summarize(rows),
	})
}

func parseDays(raw string) (int, error) {
	if raw == "" {
		return defaultDays, nil
	}
	days, err := strconv.Atoi(raw)
	if err != nil || days < 1 || days > maxDays {
		return 0, troncerrors.Invalid("days must be between 1 and 90")
	}
	return days, nil
}
