DELETE FROM sessions;

ALTER TABLE sessions ADD COLUMN plex_token text NOT NULL;

ALTER TABLE settings
    DROP COLUMN plex_token,
    DROP COLUMN plex_token_expires_at;
