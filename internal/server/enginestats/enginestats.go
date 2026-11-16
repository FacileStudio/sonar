// Package enginestats counts, per UTC day, search kind and engine, how each
// engine asked on sonar's own keys turned out. It stores numbers and nothing
// else: no query, no URL, no user and no row per search.
package enginestats

import (
	"context"
	"database/sql"
	"time"
)

// Outcome is how one engine's turn in a search ended.
type Outcome string

// The outcomes. Answered means the search stopped on this engine's results.
// Short means it returned results, but too few, and a later engine completed
// the answer. Failed covers an error, a block, a captcha and an empty answer,
// which the dispatcher does not tell apart. TimedOut means it ran out of time.
const (
	Answered Outcome = "answered"
	Short    Outcome = "short"
	Failed   Outcome = "failed"
	TimedOut Outcome = "timed_out"
)

// Attempt is one engine's turn in one search.
type Attempt struct {
	Engine  string
	Outcome Outcome
}

// Row is the counters of one engine for one kind of search on one UTC day.
// Tried is the sum of the four outcomes.
type Row struct {
	Day      string `json:"day"`
	Kind     string `json:"kind"`
	Engine   string `json:"engine"`
	Free     bool   `json:"free"`
	Tried    int64  `json:"tried"`
	Answered int64  `json:"answered"`
	Short    int64  `json:"short"`
	Failed   int64  `json:"failed"`
	TimedOut int64  `json:"timed_out"`
}

// Store is the counter over the engine_stats table.
type Store struct {
	DB  *sql.DB
	Now func() time.Time
}

func (s *Store) today() time.Time {
	now := time.Now
	if s.Now != nil {
		now = s.Now
	}
	return now().UTC().Truncate(24 * time.Hour)
}

// Record adds the attempts of one search of the given kind to today's
// counters, in one statement.
func (s *Store) Record(ctx context.Context, kind string, attempts []Attempt) error {
	if len(attempts) == 0 {
		return nil
	}
	engines, outcomes := make([]string, len(attempts)), make([]string, len(attempts))
	for i, a := range attempts {
		engines[i], outcomes[i] = a.Engine, string(a.Outcome)
	}
	_, err := s.DB.ExecContext(ctx, `
		INSERT INTO engine_stats (day, kind, engine, tried, answered, short, failed, timed_out)
		SELECT $1, $2, a.engine, count(*),
			count(*) FILTER (WHERE a.outcome = 'answered'), count(*) FILTER (WHERE a.outcome = 'short'),
			count(*) FILTER (WHERE a.outcome = 'failed'), count(*) FILTER (WHERE a.outcome = 'timed_out')
		FROM unnest($3::text[], $4::text[]) AS a(engine, outcome) GROUP BY a.engine
		ON CONFLICT (day, kind, engine) DO UPDATE SET
			tried = engine_stats.tried + EXCLUDED.tried, answered = engine_stats.answered + EXCLUDED.answered,
			short = engine_stats.short + EXCLUDED.short, failed = engine_stats.failed + EXCLUDED.failed,
			timed_out = engine_stats.timed_out + EXCLUDED.timed_out`,
		s.today(), kind, engines, outcomes)
	return err
}

// Read returns the counters of the last days UTC days, today included, newest
// day first. Free is left for the caller to fill.
func (s *Store) Read(ctx context.Context, days int) ([]Row, error) {
	rows, err := s.DB.QueryContext(ctx, `
		SELECT to_char(day, 'YYYY-MM-DD'), kind, engine, tried, answered, short, failed, timed_out
		FROM engine_stats WHERE day >= $1 ORDER BY day DESC, kind, engine`, s.Since(days))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []Row{}
	for rows.Next() {
		var r Row
		if err := rows.Scan(&r.Day, &r.Kind, &r.Engine, &r.Tried, &r.Answered, &r.Short, &r.Failed, &r.TimedOut); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// Since is the first UTC day of a period of days days that ends today.
func (s *Store) Since(days int) time.Time {
	return s.today().AddDate(0, 0, 1-days)
}
