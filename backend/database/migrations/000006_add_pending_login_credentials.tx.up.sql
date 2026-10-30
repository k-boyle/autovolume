DELETE FROM pending_logins;

ALTER TABLE pending_logins
    ADD COLUMN client_id text  NOT NULL,
    ADD COLUMN key_seed  bytea NOT NULL CHECK (octet_length(key_seed) = 32);
