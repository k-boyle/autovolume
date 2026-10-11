ALTER TABLE settings
    ADD COLUMN plex_key_seed bytea CHECK (octet_length(plex_key_seed) = 32);
