CREATE TABLE players (
    id         text PRIMARY KEY,
    name       text        NOT NULL,
    product    text        NOT NULL,
    last_seen  timestamptz NOT NULL
);

CREATE TABLE settings (
    id             boolean PRIMARY KEY DEFAULT true CHECK (id),
    plex_client_id text NOT NULL DEFAULT gen_random_uuid()::text,
    server_id      text,
    server_name    text,
    server_url     text,
    server_token   text,
    player_id      text REFERENCES players (id),
    tv_address     text,
    tv_psk         text,
    default_volume smallint CHECK (default_volume BETWEEN 0 AND 100),
    CHECK (num_nulls(server_id, server_name, server_url, server_token) IN (0, 4))
);

INSERT INTO settings DEFAULT VALUES;

CREATE TABLE pending_logins (
    id         text PRIMARY KEY,
    pin_id     bigint      NOT NULL,
    expires_at timestamptz NOT NULL
);

CREATE TABLE sessions (
    id         text PRIMARY KEY,
    username   text        NOT NULL,
    avatar_url text        NOT NULL,
    plex_token text        NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL
);

CREATE INDEX sessions_expires_at_idx ON sessions (expires_at);

CREATE TABLE volume_rules (
    server_id  text        NOT NULL,
    rating_key text        NOT NULL,
    guid       text        NOT NULL,
    type       text        NOT NULL CHECK (type IN ('movie', 'show')),
    title      text        NOT NULL,
    year       smallint,
    volume     smallint    NOT NULL CHECK (volume BETWEEN 0 AND 100),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (server_id, rating_key)
);
