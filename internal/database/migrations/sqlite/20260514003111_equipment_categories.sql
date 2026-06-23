-- +goose Up
-- +goose StatementBegin
CREATE TABLE equipment_categories (
  id TEXT PRIMARY KEY,
  -- NULL = global preseeded category visible to all orgs
  org_id TEXT REFERENCES orgs(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
  created_at INTEGER NOT NULL DEFAULT (unixepoch()),
  UNIQUE(org_id, name)
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS equipment_categories;
-- +goose StatementEnd
