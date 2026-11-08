// Package keyring keeps the engine API keys people bring themselves. Keys are
// sealed with porte's AES-256-GCM token sealer before they reach the database,
// and a key never leaves the server once saved: reads return only its hint.
package keyring

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"strings"
	"time"

	portepg "github.com/FacileStudio/porte/pg"
)

// Engines are the engines a person may bring a key for.
var Engines = []string{"brave", "exa", "firecrawl", "linkup", "parallel", "serpapi", "tavily"}

// ErrDisabled means the server has no sealing key, so it stores no user key.
var ErrDisabled = errors.New("keyring: no sealing key is configured")

// ErrUnknownEngine means the engine takes no user key.
var ErrUnknownEngine = errors.New("keyring: unknown engine")

// ErrBadKey means the key is empty, too long or carries characters no API key has.
var ErrBadKey = errors.New("keyring: malformed key")

// Saved describes one stored key without revealing it.
type Saved struct {
	Engine    string    `json:"engine"`
	Hint      string    `json:"hint"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Ring stores sealed keys in engine_keys. A Ring with a nil Sealer refuses to
// store anything and reads nothing.
type Ring struct {
	DB     *sql.DB
	Sealer *portepg.TokenSealer
}

// Enabled reports whether the Ring can seal keys.
func (r *Ring) Enabled() bool {
	return r != nil && r.Sealer != nil
}

// Validate checks an engine name and a key before anything is stored.
func Validate(engine, key string) error {
	if !slices.Contains(Engines, engine) {
		return ErrUnknownEngine
	}
	if len(key) < 8 || len(key) > 512 {
		return ErrBadKey
	}
	for _, c := range key {
		if c <= ' ' || c > '~' {
			return ErrBadKey
		}
	}
	return nil
}

// Set seals and saves key for engine, replacing any earlier one.
func (r *Ring) Set(ctx context.Context, userID int64, engine, key string) error {
	if !r.Enabled() {
		return ErrDisabled
	}
	key = strings.TrimSpace(key)
	if err := Validate(engine, key); err != nil {
		return err
	}
	sealed, err := r.Sealer.Seal(key)
	if err != nil {
		return err
	}
	_, err = r.DB.ExecContext(ctx, `INSERT INTO engine_keys (user_id, engine, sealed, hint, updated_at)
		VALUES ($1, $2, $3, $4, now())
		ON CONFLICT (user_id, engine) DO UPDATE SET sealed = EXCLUDED.sealed, hint = EXCLUDED.hint, updated_at = now()`,
		userID, engine, sealed, key[len(key)-4:])
	return err
}

// Delete forgets the key for engine. A missing key is not an error.
func (r *Ring) Delete(ctx context.Context, userID int64, engine string) error {
	_, err := r.DB.ExecContext(ctx, `DELETE FROM engine_keys WHERE user_id = $1 AND engine = $2`, userID, engine)
	return err
}
