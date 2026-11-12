package auth

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/FacileStudio/porte/directory"
	portepg "github.com/FacileStudio/porte/pg"
)

type deletionState struct {
	db *sql.DB
}

// Load returns the saved receiver state, or nil when nothing was saved yet.
func (s deletionState) Load(ctx context.Context) ([]byte, error) {
	var state []byte
	err := s.db.QueryRowContext(ctx, `SELECT state FROM deletion_state WHERE id = 1`).Scan(&state)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return state, err
}

// Save replaces the saved receiver state.
func (s deletionState) Save(ctx context.Context, state []byte) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO deletion_state (id, state) VALUES (1, $1)
		ON CONFLICT (id) DO UPDATE SET state = EXCLUDED.state, updated_at = now()`, state)
	return err
}

// eraseAccount deletes the porte user behind (provider, subject). Every sonar
// table that holds a person's data (search_usage, engine_keys, sessions,
// identities, login codes) references porte_users with ON DELETE CASCADE, so
// one statement erases it all, sealed engine keys included. The provider
// matches with or without a trailing slash, as porte/kit does, because the
// configured issuer and the stored rows have used both spellings.
func eraseAccount(db *sql.DB) func(ctx context.Context, provider, subject string) error {
	return func(ctx context.Context, provider, subject string) error {
		_, err := db.ExecContext(ctx, `DELETE FROM porte_users WHERE id IN (
			SELECT user_id FROM porte_identities WHERE rtrim(provider, '/') = rtrim($1, '/') AND subject = $2)`,
			provider, subject)
		return err
	}
}

func newDeletions(db *sql.DB, store *portepg.Store, client directory.Client, logger *slog.Logger) *directory.Deletions {
	return &directory.Deletions{
		Client:          client,
		OnPersonDeleted: eraseAccount(db),
		Erase:           store.EraseIdentity,
		Holds:           store.HoldsIdentity,
		State:           deletionState{db: db},
		Logger:          logger,
	}
}
