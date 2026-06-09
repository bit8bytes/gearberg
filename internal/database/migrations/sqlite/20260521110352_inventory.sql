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
  code INTEGER NOT NULL,
  storage_object_id TEXT DEFAULT NULL REFERENCES storage_objects(id) ON DELETE SET NULL,
  qr_object_id TEXT DEFAULT NULL REFERENCES storage_objects(id) ON DELETE SET NULL,
  purchase_price INTEGER,
  rental_price INTEGER,
  notes TEXT,
  updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
  created_at INTEGER NOT NULL DEFAULT (unixepoch()),
  weight_g INTEGER,
  width_mm INTEGER,
  height_mm INTEGER,
  depth_mm INTEGER,
  power_mw INTEGER,
  current_ma INTEGER,
  UNIQUE(org_id, code)
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS inventory;
-- +goose StatementEnd
