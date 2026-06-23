-- +goose Up
-- +goose StatementBegin
CREATE TABLE equipment_bulk_items (
  id TEXT PRIMARY KEY,
  equipment_id TEXT NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
  quantity INTEGER NOT NULL DEFAULT 1,
  purchase_price INTEGER,
  purchased_at INTEGER,
  remark TEXT,
  updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
  created_at INTEGER NOT NULL DEFAULT (unixepoch())
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS equipment_bulk_items;
-- +goose StatementEnd
