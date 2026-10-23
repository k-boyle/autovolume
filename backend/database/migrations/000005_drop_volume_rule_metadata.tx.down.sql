DELETE FROM volume_rules;

ALTER TABLE volume_rules
    ADD COLUMN guid  text     NOT NULL,
    ADD COLUMN type  text     NOT NULL CHECK (type IN ('movie', 'show')),
    ADD COLUMN title text     NOT NULL,
    ADD COLUMN year  smallint;
