-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS org_roles (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    rank INTEGER NOT NULL DEFAULT 0
) STRICT;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS org_roles;
-- +goose StatementEnd