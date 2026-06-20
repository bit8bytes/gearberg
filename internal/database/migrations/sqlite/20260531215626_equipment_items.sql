-- +goose Up
-- +goose StatementBegin
CREATE TABLE equipment_items (
  id TEXT PRIMARY KEY,
  equipment_id TEXT NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
  parent_equipment_item_id TEXT REFERENCES equipment_items(id) ON DELETE SET NULL,
  internal_id INTEGER NOT NULL,
  is_active INTEGER NOT NULL DEFAULT 1,
  quantity INTEGER NOT NULL DEFAULT 1,
  remark TEXT,
  purchase_price INTEGER,
  purchased_at INTEGER,
  next_inspection_at INTEGER,
  manufacturer_serial TEXT,
  updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
  created_at INTEGER NOT NULL DEFAULT (unixepoch())
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS equipment_items;
-- +goose StatementEnd
