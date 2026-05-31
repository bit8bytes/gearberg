-- +goose Up
-- +goose StatementBegin
CREATE TABLE org_settings (
  id TEXT PRIMARY KEY,
  org_id TEXT NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
  currency TEXT NOT NULL,
  vat_rate REAL NOT NULL,
  timezone TEXT NOT NULL,
  updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
  created_at INTEGER NOT NULL DEFAULT (unixepoch())
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS org_settings;
-- +goose StatementEnd
