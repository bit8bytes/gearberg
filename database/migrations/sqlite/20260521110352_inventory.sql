-- +goose Up
-- +goose StatementBegin
CREATE TABLE inventory (
  id TEXT PRIMARY KEY,
  org_id TEXT NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
  category_id TEXT NOT NULL REFERENCES equipment_categories(id) ON DELETE RESTRICT,
  manufacturer_id TEXT REFERENCES manufacturers(id) ON DELETE RESTRICT,
  type_id INTEGER NOT NULL REFERENCES inventory_types(id) ON DELETE RESTRICT,
  usage_type_id INTEGER NOT NULL REFERENCES usage_types(id) ON DELETE RESTRICT,
  name TEXT NOT NULL,
  code TEXT,
  storage_object_id TEXT DEFAULT NULL REFERENCES storage_objects(id) ON DELETE SET NULL,
  total_stock INTEGER NOT NULL DEFAULT 1,
  purchase_price INTEGER,
  rental_price INTEGER,
  notes TEXT,
  updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
  created_at INTEGER NOT NULL DEFAULT (unixepoch())
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS inventory;
-- +goose StatementEnd
