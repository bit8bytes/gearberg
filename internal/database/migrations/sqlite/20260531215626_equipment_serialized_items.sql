-- +goose Up
-- +goose StatementBegin
CREATE TABLE equipment_serialized_items (
  id TEXT PRIMARY KEY,
  org_id TEXT NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
  equipment_id TEXT NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
  parent_item_id TEXT REFERENCES equipment_serialized_items(id) ON DELETE SET NULL,
  serial_number TEXT NOT NULL,
  code TEXT,
  is_active INTEGER NOT NULL DEFAULT 1,
  remark TEXT,
  purchase_price INTEGER,
  purchased_at INTEGER,
  next_inspection_at INTEGER,
  manufacturer_serial TEXT,
  updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
  created_at INTEGER NOT NULL DEFAULT (unixepoch()),
  UNIQUE (org_id, serial_number)
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS equipment_serialized_items;
-- +goose StatementEnd
