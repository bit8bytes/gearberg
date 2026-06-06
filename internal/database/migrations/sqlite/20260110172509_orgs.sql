-- +goose Up
-- +goose StatementBegin
CREATE TABLE orgs (
  id TEXT PRIMARY KEY,
  name TEXT UNIQUE NOT NULL,
  updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
  created_at INTEGER NOT NULL DEFAULT (unixepoch())
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS orgs;
-- +goose StatementEnd