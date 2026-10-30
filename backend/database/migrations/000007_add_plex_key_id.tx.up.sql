DELETE FROM pending_logins;

ALTER TABLE pending_logins ADD COLUMN key_id text NOT NULL;

ALTER TABLE settings ADD COLUMN plex_key_id text;

UPDATE settings SET plex_key_id = 'autovolume' WHERE plex_key_seed IS NOT NULL;
