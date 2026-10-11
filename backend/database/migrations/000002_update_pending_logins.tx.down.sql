DROP TABLE pending_logins;

CREATE TABLE pending_logins (
    id         text PRIMARY KEY,
    pin_id     bigint      NOT NULL,
    expires_at timestamptz NOT NULL
);
