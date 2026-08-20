-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS provider_types (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS provider_types;
-- +goose StatementEnd
