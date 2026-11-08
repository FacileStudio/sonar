package keyring

import (
	"context"
	"strings"
)

// List describes the keys userID saved, in engine order.
func (r *Ring) List(ctx context.Context, userID int64) ([]Saved, error) {
	saved := []Saved{}
	if !r.Enabled() {
		return saved, nil
	}
	rows, err := r.DB.QueryContext(ctx,
		`SELECT engine, hint, updated_at FROM engine_keys WHERE user_id = $1 ORDER BY engine`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var s Saved
		if err := rows.Scan(&s.Engine, &s.Hint, &s.UpdatedAt); err != nil {
			return nil, err
		}
		saved = append(saved, s)
	}
	return saved, rows.Err()
}

// Keys opens every key userID saved, by engine. A value that is not sealed, or
// that no configured key opens, is skipped rather than used.
func (r *Ring) Keys(ctx context.Context, userID int64) (map[string]string, error) {
	keys := map[string]string{}
	if !r.Enabled() {
		return keys, nil
	}
	rows, err := r.DB.QueryContext(ctx, `SELECT engine, sealed FROM engine_keys WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var engine, sealed string
		if err := rows.Scan(&engine, &sealed); err != nil {
			return nil, err
		}
		if !strings.HasPrefix(sealed, "pv1:") {
			continue
		}
		if key, err := r.Sealer.Open(sealed); err == nil && key != "" {
			keys[engine] = key
		}
	}
	return keys, rows.Err()
}
