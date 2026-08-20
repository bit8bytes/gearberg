-- +goose Up
-- +goose StatementBegin
CREATE TABLE equipment (
  id TEXT PRIMARY KEY,
  org_id TEXT NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
  equipment_type_id INTEGER NOT NULL DEFAULT 1 REFERENCES equipment_types(id) ON DELETE RESTRICT,
  tracking_type_id INTEGER REFERENCES tracking_types(id) ON DELETE RESTRICT,
  category_id TEXT REFERENCES equipment_categories(id) ON DELETE RESTRICT,
  manufacturer_id TEXT REFERENCES equipment_manufacturers(id) ON DELETE RESTRICT,
  usage_type_id INTEGER NOT NULL REFERENCES usage_types(id) ON DELETE RESTRICT,
  location_id TEXT REFERENCES warehouse_locations(id) ON DELETE SET NULL,
  storage_object_id TEXT REFERENCES storage_objects(id) ON DELETE SET NULL,
  name TEXT NOT NULL,
  is_archived INTEGER NOT NULL DEFAULT 0,
  rental_price INTEGER,
  resale_price INTEGER,
  notes TEXT,
  weight_g INTEGER,
  width_mm INTEGER,
  height_mm INTEGER,
  depth_mm INTEGER,
  voltage_mv INTEGER,
  current_ma INTEGER,
  power_mw INTEGER,
  wire_gauge_mm2_x100 INTEGER,
  updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
  created_at INTEGER NOT NULL DEFAULT (unixepoch())
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS equipment;
-- +goose StatementEnd
