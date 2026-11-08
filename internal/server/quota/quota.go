// Package quota counts searches per user and UTC day. It stores a number and
// nothing else: no query, no engine, no time of day.
package quota

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// Counter is the daily search counter over the search_usage table.
type Counter struct {
	DB  *sql.DB
	Now func() time.Time
}

func (c *Counter) today() time.Time {
	now := time.Now
	if c.Now != nil {
		now = c.Now
	}
	return now().UTC().Truncate(24 * time.Hour)
}

// Take counts one search for userID against limit and reports whether it fit.
// A negative limit is unlimited and still counted; a zero limit never fits.
// The check and the increment are one statement, so concurrent searches
// cannot overshoot the limit.
func (c *Counter) Take(ctx context.Context, userID, limit int64) (bool, error) {
	if limit == 0 {
		return false, nil
	}
	var count int64
	err := c.DB.QueryRowContext(ctx, `
		INSERT INTO search_usage (user_id, day, count) VALUES ($1, $2, 1)
		ON CONFLICT (user_id, day) DO UPDATE SET count = search_usage.count + 1
		WHERE $3 < 0 OR search_usage.count < $3
		RETURNING count`, userID, c.today(), limit).Scan(&count)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

// Refund gives back one search, for a search no engine could answer.
func (c *Counter) Refund(ctx context.Context, userID int64) error {
	_, err := c.DB.ExecContext(ctx, `
		UPDATE search_usage SET count = count - 1
		WHERE user_id = $1 AND day = $2 AND count > 0`, userID, c.today())
	return err
}

// Used answers how many searches userID made today.
func (c *Counter) Used(ctx context.Context, userID int64) (int64, error) {
	var count int64
	err := c.DB.QueryRowContext(ctx, `SELECT count FROM search_usage WHERE user_id = $1 AND day = $2`,
		userID, c.today()).Scan(&count)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return count, err
}

// Sweep deletes the counters of days before yesterday, which no quota reads.
func (c *Counter) Sweep(ctx context.Context) (int64, error) {
	result, err := c.DB.ExecContext(ctx, `DELETE FROM search_usage WHERE day < $1`, c.today().AddDate(0, 0, -1))
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// ResetsAt is when today's counters stop applying: the next UTC midnight.
func (c *Counter) ResetsAt() time.Time {
	return c.today().AddDate(0, 0, 1)
}
