-- +goose Up
-- engine_stats counts, per UTC day, search kind and engine, how each engine
-- asked on sonar's own keys turned out. It holds counters and nothing else:
-- no query, no URL, no user and no row per search.

CREATE TABLE engine_stats (
	day       date NOT NULL,
	kind      text NOT NULL,
	engine    text NOT NULL,
	tried     bigint NOT NULL DEFAULT 0,
	answered  bigint NOT NULL DEFAULT 0,
	short     bigint NOT NULL DEFAULT 0,
	failed    bigint NOT NULL DEFAULT 0,
	timed_out bigint NOT NULL DEFAULT 0,
	PRIMARY KEY (day, kind, engine)
);

-- +goose Down
DROP TABLE engine_stats;
