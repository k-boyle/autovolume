ALTER TABLE settings
    ADD COLUMN plex_token            text,
    ADD COLUMN plex_token_expires_at timestamptz,
    ADD CHECK (num_nulls(plex_token, plex_token_expires_at) IN (0, 2));

ALTER TABLE sessions DROP COLUMN plex_token;
