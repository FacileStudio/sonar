package server

import (
	"context"
	"database/sql"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/sonar/internal/server/imgsig"
	"github.com/FacileStudio/sonar/internal/server/keyring"
	"github.com/FacileStudio/sonar/internal/server/modules/auth"
	"github.com/FacileStudio/sonar/internal/server/modules/search"
	"github.com/FacileStudio/sonar/internal/server/modules/summary"
	"github.com/FacileStudio/sonar/internal/server/plans"
	"github.com/FacileStudio/sonar/internal/server/quota"
	"github.com/FacileStudio/sonar/websearch"
)

type app struct {
	kits    *auth.Kits
	search  *search.Service
	suggest *search.Suggester
	summary *summary.Service
	ring    *keyring.Ring
	signer  *imgsig.Signer
	counter *quota.Counter
	log     *slog.Logger
}

func assemble(ctx context.Context, db *sql.DB, cfg Config, log *slog.Logger) (*app, error) {
	gate := plans.New(db, cfg.OIDC.Issuer, log)
	if err := gate.Load(ctx); err != nil {
		return nil, err
	}
	kits, err := auth.New(ctx, db, auth.Config{Porte: cfg.OIDC, ClientID: cfg.OIDC.ClientID,
		ClientSecret: cfg.ClientSecret, Sealer: cfg.Sealer}, gate.Apply, log)
	if err != nil {
		return nil, err
	}
	instance, err := websearch.New(websearch.Config{Keys: cfg.Engines.Keys, Searxng: cfg.Engines.Searxng,
		Timeout: 12 * time.Second})
	if err != nil {
		return nil, err
	}
	signer, err := imgsig.New(cfg.ImageKey)
	if err != nil {
		return nil, err
	}
	ring := &keyring.Ring{DB: db, Sealer: cfg.Sealer}
	counter := &quota.Counter{DB: db}
	a := &app{kits: kits, ring: ring, signer: signer, counter: counter, log: log,
		suggest: search.NewSuggester(cfg.SuggestURL)}
	a.search = &search.Service{Instance: instance, ForKeys: forKeys, Plans: gate, Quota: counter, Keys: ring,
		Signer: signer, Logger: log, Web: cfg.Engines.Web.Within(instance.Engines()),
		Sources: cfg.Engines.Sources.Within(instance.Engines())}
	a.wireSummary(db, gate, cfg)
	log.Info("search engines ready", slog.Any("engines", instance.Engines()),
		slog.Any("web_chain", a.search.Web.Engines), slog.Any("summary_engines", a.search.Sources.Engines))
	return a, nil
}

func (a *app) wireSummary(db *sql.DB, access summary.Access, cfg Config) {
	a.summary = &summary.Service{DB: db, Access: access, Sources: a.summarySources}
	if writer := summary.NewBulle(cfg.BulleURL, cfg.BulleKey); cfg.Summary && writer != nil {
		a.summary.Writer = writer
	}
}

func (a *app) summarySources(ctx context.Context, userID int64, query string, count int) ([]summary.Source, error) {
	hits, err := a.search.SummarySources(ctx, userID, query, count)
	out := make([]summary.Source, 0, len(hits))
	for _, h := range hits {
		out = append(out, summary.Source{Title: h.Title, URL: h.URL, Snippet: h.Snippet})
	}
	return out, err
}

func forKeys(keys map[string]string) (search.Engines, error) {
	return websearch.New(websearch.Config{Keys: keys, Timeout: 12 * time.Second})
}

func (a *app) background(ctx context.Context) {
	a.kits.Run(ctx)
	go func() {
		ticker := time.NewTicker(6 * time.Hour)
		defer ticker.Stop()
		for {
			if _, err := a.counter.Sweep(ctx); err != nil && ctx.Err() == nil {
				a.log.Warn("quota sweep failed", slog.Any("error", err))
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

func userKey(r *http.Request) string {
	if identity, ok := porte.From(r.Context()); ok && identity.UserID > 0 {
		return "user:" + strconv.FormatInt(identity.UserID, 10)
	}
	return "anon:" + r.RemoteAddr
}
