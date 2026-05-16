-- +goose Up
-- +goose StatementBegin
CREATE TABLE company_settings (
  id TEXT PRIMARY KEY,
  company_id TEXT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
  currency TEXT NOT NULL,
  vat_rate REAL NOT NULL,
  timezone TEXT NOT NULL,
  updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
  created_at INTEGER NOT NULL DEFAULT (unixepoch())
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS company_settings;
-- +goose StatementEnd
