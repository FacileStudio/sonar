package server

import (
	"log/slog"

	"github.com/FacileStudio/porte/exchange"
	"github.com/FacileStudio/sonar/internal/server/modules/summary"
)

func (a *app) wireSummary(access summary.Access, cfg Config) error {
	a.summary = &summary.Service{Access: access, Sources: a.summarySources}
	if !cfg.Summary || cfg.BulleURL == "" || cfg.OIDC.Issuer == "" {
		return nil
	}
	exchanger, err := exchange.New(exchange.Config{
		Issuer:       cfg.OIDC.Issuer,
		ClientID:     cfg.OIDC.ClientID,
		ClientSecret: cfg.ClientSecret,
		Identities:   a.kits.Identities,
		Logger:       a.log.With(slog.String("module", "summary")),
	})
	if err != nil {
		return err
	}
	source := exchanger.Source(exchange.Target{Audience: "bulle", Scopes: []string{"bulle.read", "bulle.write"}})
	writer, err := summary.NewBulle(cfg.BulleURL, source.Token, source.Forget)
	if err != nil {
		return err
	}
	a.summary.Writer = writer
	return nil
}
