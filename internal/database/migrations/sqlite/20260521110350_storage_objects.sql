-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS storage_objects (
  id TEXT PRIMARY KEY,
  org_id TEXT NOT NULL,
  key TEXT UNIQUE NOT NULL,
  backend TEXT NOT NULL,
  filename TEXT NOT NULL,
  content_type TEXT NOT NULL,
  size INTEGER NOT NULL,
  encryption_key_id INTEGER NOT NULL DEFAULT 0,
  created_at INTEGER NOT NULL DEFAULT(unixepoch())
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS storage_objects;
-- +goose StatementEnd