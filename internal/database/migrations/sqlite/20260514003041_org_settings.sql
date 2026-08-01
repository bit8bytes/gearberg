-- +goose Up
-- +goose StatementBegin
CREATE TABLE org_settings (
  id TEXT PRIMARY KEY,
  org_id TEXT NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
  currency TEXT NOT NULL DEFAULT 'EUR',
  vat_rate INTEGER NOT NULL DEFAULT 1900,
  timezone TEXT NOT NULL DEFAULT 'Europe/Berlin',
  updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
  created_at INTEGER NOT NULL DEFAULT (unixepoch())
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS org_settings;
-- +goose StatementEnd
