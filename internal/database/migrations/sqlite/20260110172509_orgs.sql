-- +goose Up
-- +goose StatementBegin
CREATE TABLE orgs (
  id TEXT PRIMARY KEY,
  display_name TEXT NOT NULL,
  updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
  created_at INTEGER NOT NULL DEFAULT (unixepoch())
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS orgs;
-- +goose StatementEnd