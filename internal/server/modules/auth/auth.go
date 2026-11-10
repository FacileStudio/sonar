// Package auth is sonar's sign-in, delegated to porte against Registre, plus
// the two background links to Registre: the directory sync that carries plan
// entitlements and the receiver that erases people Registre erased.
package auth

import (
	"context"
	"database/sql"
	"log/slog"
	"strings"

	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/directory"
	"github.com/FacileStudio/porte/oidc"
	portepg "github.com/FacileStudio/porte/pg"
	"github.com/FacileStudio/porte/session"
)

// Config is what sign-in needs from the process configuration.
type Config struct {
	Porte        porte.Config
	ClientID     string
	ClientSecret string
	Sealer       *portepg.TokenSealer
}

// Kits is porte, assembled.
type Kits struct {
	DB         *sql.DB
	Sessions   *session.Manager
	OIDC       *oidc.Kit
	Syncer     *directory.Syncer
	Deletions  *directory.Deletions
	Identities porte.IdentityStore
	Issuer     string
}

type pictureUsers struct {
	*portepg.UserStore
}

// UpsertFromOIDC copies Registre's picture claim to AvatarURL before upserting.
func (u pictureUsers) UpsertFromOIDC(ctx context.Context, claims porte.Claims) (int64, error) {
	if claims.AvatarURL == "" && strings.HasPrefix(claims.Picture, "https://") {
		claims.AvatarURL = claims.Picture
	}
	return u.UserStore.UpsertFromOIDC(ctx, claims)
}

func backfillPictures(ctx context.Context, db *sql.DB, issuer string) (int64, error) {
	if !strings.HasPrefix(issuer, "https://") {
		return 0, nil
	}
	base := strings.TrimRight(issuer, "/")
	res, err := db.ExecContext(ctx, `UPDATE porte_users
        SET avatar_url = $1 || '/avatars/' || porte_identities.subject,
            avatar_source = 'oidc'
       FROM porte_identities
      WHERE porte_identities.user_id = porte_users.id
        AND rtrim(porte_identities.provider, '/') = $2
        AND porte_users.avatar_url = ''`, base, base)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// New builds sessions and runs OIDC discovery. apply receives every directory
// snapshot. Without an issuer there is no Syncer and no Deletions.
func New(ctx context.Context, db *sql.DB, cfg Config, apply func(context.Context, directory.Snapshot) error,
	logger *slog.Logger) (*Kits, error) {
	store := portepg.New(db, portepg.WithTokenSealer(cfg.Sealer))
	if cfg.Sealer != nil {
		resealInBackground(ctx, store, logger)
	}
	sessions, err := session.New(cfg.Porte, session.Deps{Sessions: store.Sessions(), Logger: logger})
	if err != nil {
		return nil, err
	}
	kit, err := oidc.New(ctx, cfg.Porte, oidc.Deps{
		Users: pictureUsers{UserStore: store.Users()}, Identities: store.Identities(), Sessions: sessions, Codes: store.LoginCodes(), Logger: logger,
	})
	if err != nil {
		return nil, err
	}
	kits := &Kits{DB: db, Sessions: sessions, OIDC: kit, Identities: store.Identities(), Issuer: cfg.Porte.Issuer}
	if cfg.Porte.Issuer != "" {
		if strings.HasPrefix(cfg.Porte.Issuer, "https://") {
			if _, err := backfillPictures(ctx, db, cfg.Porte.Issuer); err != nil && logger != nil {
				logger.Warn("backfill pictures failed", slog.Any("error", err))
			}
		}
		client := directory.Client{Issuer: cfg.Porte.Issuer, ClientID: cfg.ClientID, ClientSecret: cfg.ClientSecret}
		kits.Syncer = &directory.Syncer{Client: client, Apply: apply, Logger: logger}
		kits.Deletions = newDeletions(db, store, client, logger)
	}
	return kits, nil
}

func resealInBackground(ctx context.Context, store *portepg.Store, logger *slog.Logger) {
	go func() {
		count, err := store.Identities().ResealIdentities(ctx)
		if err != nil && ctx.Err() == nil {
			logger.Warn("porte token reseal failed", slog.Int("resealed", count), slog.Any("error", err))
			return
		}
		logger.Info("porte token reseal complete", slog.Int("resealed", count))
	}()
}

// Run keeps the directory in sync and polls for erased people until ctx ends.
// Run it in exactly one replica.
func (k *Kits) Run(ctx context.Context) {
	if k.Syncer != nil {
		go k.Syncer.Run(ctx)
	}
	if k.Deletions != nil {
		go k.Deletions.Run(ctx)
	}
}
