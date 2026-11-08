package plans

import (
	"context"

	"github.com/FacileStudio/porte/forfait"
)

// Daily answers the daily search quota of the porte user: the most generous
// value across the user's Registre identities, -1 for unlimited. Before the
// Gate is authoritative it answers DefaultDaily.
func (g *Gate) Daily(ctx context.Context, userID int64) (int64, error) {
	if !g.Active() {
		return DefaultDaily, nil
	}
	subjects, err := g.subjects(ctx, userID)
	if err != nil {
		return 0, err
	}
	var best int64
	for _, subject := range subjects {
		limit := g.store.Limit(forfait.Scope{Subject: subject}, SearchesDaily)
		if limit < 0 {
			return -1, nil
		}
		best = max(best, limit)
	}
	return best, nil
}

// CheckAI answers nil when the user's plan includes ai.access and the 402
// refusal otherwise. It refuses until the Gate is authoritative.
func (g *Gate) CheckAI(ctx context.Context, userID int64) error {
	refusal := g.Refusal("AI summaries need a paid plan.", ErrAIAccess)
	if !g.Active() {
		return refusal
	}
	subjects, err := g.subjects(ctx, userID)
	if err != nil {
		return err
	}
	for _, subject := range subjects {
		if g.store.Enabled(forfait.Scope{Subject: subject}, AIAccess) {
			return nil
		}
	}
	return refusal
}

func (g *Gate) subjects(ctx context.Context, userID int64) ([]string, error) {
	rows, err := g.db.QueryContext(ctx, `SELECT subject FROM porte_identities WHERE user_id = $1 AND provider = $2`,
		userID, g.issuer)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var subjects []string
	for rows.Next() {
		var subject string
		if err := rows.Scan(&subject); err != nil {
			return nil, err
		}
		subjects = append(subjects, subject)
	}
	if len(subjects) == 0 {
		subjects = []string{""}
	}
	return subjects, rows.Err()
}
