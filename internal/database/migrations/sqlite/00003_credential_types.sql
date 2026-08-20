-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS credential_types (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS credential_types;
-- +goose StatementEnd