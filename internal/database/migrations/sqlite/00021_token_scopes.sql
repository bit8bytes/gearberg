-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS token_scopes (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS token_scopes;
-- +goose StatementEnd
