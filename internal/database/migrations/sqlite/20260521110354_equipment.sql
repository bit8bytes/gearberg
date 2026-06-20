-- +goose Up
-- +goose StatementBegin
CREATE TABLE equipment (
  id TEXT PRIMARY KEY,
  org_id TEXT NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
  tracking_type_id INTEGER REFERENCES tracking_types(id) ON DELETE RESTRICT,
  category_id TEXT REFERENCES equipment_categories(id) ON DELETE RESTRICT,
  manufacturer_id TEXT REFERENCES equipment_manufacturers(id) ON DELETE RESTRICT,
  usage_type_id INTEGER NOT NULL REFERENCES usage_types(id) ON DELETE RESTRICT,
  location_id TEXT REFERENCES warehouse_locations(id) ON DELETE SET NULL,
  storage_object_id TEXT REFERENCES storage_objects(id) ON DELETE SET NULL,
  name TEXT NOT NULL,
  has_content INTEGER NOT NULL DEFAULT 0,
  rental_price INTEGER,
  resale_price INTEGER,
  notes TEXT,
  weight_g INTEGER,
  width_mm INTEGER,
  height_mm INTEGER,
  depth_mm INTEGER,
  voltage_v INTEGER,
  current_ma INTEGER,
  power_mw INTEGER,
  updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
  created_at INTEGER NOT NULL DEFAULT (unixepoch())
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS equipment;
-- +goose StatementEnd
