// Package plans reads the entitlements Registre serves for sonar: the daily
// search quota and access to the AI features. It feeds every directory
// snapshot to a porte/forfait Store and keeps the Store's state in one row, so
// a restart never forgets the plans it last saw.
package plans

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log/slog"
	"sync/atomic"

	"github.com/FacileStudio/porte/directory"
	"github.com/FacileStudio/porte/forfait"
	troncerrors "github.com/FacileStudio/tronc/errors"
)

const (
	// SearchesDaily is the number of searches a person may run per UTC day on
	// sonar's own engine keys. -1 is unlimited.
	SearchesDaily = "sonar.searches_daily"
	// AIAccess turns the paid AI features on, the suite-wide flag.
	AIAccess = "ai.access"
	// DefaultDaily is the free plan's quota, and the quota everyone gets until
	// Registre has served an entitlements snapshot.
	DefaultDaily = 20
)

// ErrAIAccess is the cause of every AI refusal.
var ErrAIAccess = errors.New("plans: ai.access is not part of this plan")

// Gate is sonar's view of the plans. Unlike forfait's default it never fails
// open: until an authoritative snapshot arrives every person reads the free
// plan, because an open quota spends real engine credits.
type Gate struct {
	db     *sql.DB
	issuer string
	logger *slog.Logger
	store  *forfait.Store
	active atomic.Bool
}

// New builds the Gate for the OIDC issuer people sign in with.
func New(db *sql.DB, issuer string, logger *slog.Logger) *Gate {
	store := forfait.NewStore(
		forfait.Feature{Key: SearchesDaily, Kind: forfait.Limit, Default: json.RawMessage("20")},
		forfait.Feature{Key: AIAccess, Kind: forfait.Bool, Default: json.RawMessage("false")},
	)
	store.Issuer = issuer
	store.Logger = logger
	return &Gate{db: db, issuer: issuer, logger: logger, store: store}
}

// Load restores the state the last Apply saved. A row that cannot be imported
// keeps the Gate at the free defaults until the next snapshot.
func (g *Gate) Load(ctx context.Context) error {
	var document string
	err := g.db.QueryRowContext(ctx, `SELECT document FROM forfait_state WHERE id = 1`).Scan(&document)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := g.store.Import([]byte(document)); err != nil {
		g.logger.Error("plans: saved entitlements could not be restored, enforcing defaults", slog.Any("error", err))
		return nil
	}
	g.refresh([]byte(document))
	return nil
}

// Apply feeds one directory snapshot to the Store and saves the result. It is
// the directory Syncer's Apply, so a failed save makes the Syncer retry.
func (g *Gate) Apply(ctx context.Context, snapshot directory.Snapshot) error {
	g.store.Apply(snapshot)
	document, err := g.store.Export()
	if err != nil {
		return err
	}
	if _, err := g.db.ExecContext(ctx, `INSERT INTO forfait_state (id, document, updated_at) VALUES (1, $1, now())
		ON CONFLICT (id) DO UPDATE SET document = EXCLUDED.document, updated_at = now()`, string(document)); err != nil {
		return err
	}
	g.refresh(document)
	return nil
}

func (g *Gate) refresh(document []byte) {
	var state struct {
		Revision int64 `json:"entitlements_revision"`
	}
	if json.Unmarshal(document, &state) == nil && state.Revision > 0 {
		g.active.Store(true)
	}
}

// Active reports whether entitlements are authoritative.
func (g *Gate) Active() bool {
	return g != nil && g.active.Load()
}

// UpgradeURL is the Registre page where a person changes their plan.
func (g *Gate) UpgradeURL() string {
	return forfait.UpgradeURL(g.issuer, "")
}

// Refusal is the 402 envelope error for a refused paid feature.
func (g *Gate) Refusal(message string, cause error) error {
	return troncerrors.New("payment_required", message, cause).WithUpgradeURL(g.UpgradeURL())
}
