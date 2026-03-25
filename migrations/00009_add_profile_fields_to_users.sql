-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS full_name   VARCHAR(255),
    ADD COLUMN IF NOT EXISTS birth_date  DATE,
    ADD COLUMN IF NOT EXISTS bio         TEXT,
    ADD COLUMN IF NOT EXISTS phone       VARCHAR(50),
    ADD COLUMN IF NOT EXISTS city        VARCHAR(100);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users
    DROP COLUMN IF EXISTS full_name,
    DROP COLUMN IF EXISTS birth_date,
    DROP COLUMN IF EXISTS bio,
    DROP COLUMN IF EXISTS phone,
    DROP COLUMN IF EXISTS city;
-- +goose StatementEnd
