-- +goose Up
-- porte owns sign-in: users, Registre identities, sessions and the one-time
-- codes that carry a browser login to a CLI. The tables are porte/pg Schema
-- as of porte v0.13.0, written out for a fresh database.

CREATE TABLE porte_users (
	id             bigserial PRIMARY KEY,
	facile_id      text UNIQUE,
	email          text NOT NULL UNIQUE,
	email_verified boolean NOT NULL DEFAULT false,
	name           text NOT NULL DEFAULT '',
	avatar_url     text NOT NULL DEFAULT '',
	avatar_source  text NOT NULL DEFAULT '',
	locale         text NOT NULL DEFAULT 'en',
	created_at     timestamptz NOT NULL DEFAULT now(),
	updated_at     timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE porte_identities (
	user_id         bigint NOT NULL REFERENCES porte_users(id) ON DELETE CASCADE,
	provider        text NOT NULL,
	subject         text NOT NULL,
	password_hash   text NOT NULL DEFAULT '',
	access_token    text NOT NULL DEFAULT '',
	refresh_token   text NOT NULL DEFAULT '',
	token_expiry    timestamptz,
	roles           jsonb,
	roles_synced_at timestamptz,
	synced_at       timestamptz,
	created_at      timestamptz DEFAULT now(),
	PRIMARY KEY (provider, subject)
);
CREATE INDEX porte_identities_user_idx ON porte_identities (user_id);

CREATE TABLE porte_sessions (
	id           bigserial PRIMARY KEY,
	token_hash   text NOT NULL UNIQUE,
	user_id      bigint NOT NULL REFERENCES porte_users(id) ON DELETE CASCADE,
	label        text NOT NULL DEFAULT '',
	created_at   timestamptz NOT NULL DEFAULT now(),
	last_used_at timestamptz NOT NULL DEFAULT now(),
	expires_at   timestamptz
);
CREATE INDEX porte_sessions_user_idx ON porte_sessions (user_id);
CREATE INDEX porte_sessions_expiry_idx ON porte_sessions (expires_at) WHERE expires_at IS NOT NULL;

CREATE TABLE porte_login_codes (
	code_hash   text PRIMARY KEY,
	user_id     bigint NOT NULL REFERENCES porte_users(id) ON DELETE CASCADE,
	expires_at  timestamptz NOT NULL,
	consumed_at timestamptz,
	provider    text NOT NULL DEFAULT '',
	subject     text NOT NULL DEFAULT ''
);

CREATE TABLE porte_magic_links (
	token_hash  text PRIMARY KEY,
	user_id     bigint NOT NULL REFERENCES porte_users(id) ON DELETE CASCADE,
	ip          text NOT NULL DEFAULT '',
	created_at  timestamptz NOT NULL,
	expires_at  timestamptz NOT NULL,
	consumed_at timestamptz
);
CREATE INDEX porte_magic_links_user_idx ON porte_magic_links (user_id, created_at);
CREATE INDEX porte_magic_links_expiry_idx ON porte_magic_links (expires_at);

-- +goose Down
DROP TABLE porte_magic_links;
DROP TABLE porte_login_codes;
DROP TABLE porte_sessions;
DROP TABLE porte_identities;
DROP TABLE porte_users;
