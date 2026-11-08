-- +goose Up
-- sonar keeps no search history. search_usage holds one counter per user and
-- UTC day for the daily quota, and nothing about what was searched.
-- engine_keys holds a user's own engine API keys, sealed with porte's
-- AES-256-GCM sealer; hint is the last four characters, for display only.
-- forfait_state persists the plan entitlements Registre serves, and
-- deletion_state the GDPR receiver's cursor.

CREATE TABLE search_usage (
	user_id bigint NOT NULL REFERENCES porte_users(id) ON DELETE CASCADE,
	day     date NOT NULL,
	count   integer NOT NULL DEFAULT 0,
	PRIMARY KEY (user_id, day)
);
CREATE INDEX search_usage_day_idx ON search_usage (day);

CREATE TABLE engine_keys (
	user_id    bigint NOT NULL REFERENCES porte_users(id) ON DELETE CASCADE,
	engine     text NOT NULL,
	sealed     text NOT NULL,
	hint       text NOT NULL DEFAULT '',
	updated_at timestamptz NOT NULL DEFAULT now(),
	PRIMARY KEY (user_id, engine)
);

CREATE TABLE forfait_state (
	id         smallint PRIMARY KEY CHECK (id = 1),
	document   text NOT NULL,
	updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE deletion_state (
	id         smallint PRIMARY KEY CHECK (id = 1),
	state      bytea NOT NULL,
	updated_at timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE deletion_state;
DROP TABLE forfait_state;
DROP TABLE engine_keys;
DROP TABLE search_usage;
