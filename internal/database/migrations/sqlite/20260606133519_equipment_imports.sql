-- +goose Up
-- +goose StatementBegin
CREATE TABLE equipment_imports (
  id TEXT NOT NULL PRIMARY KEY,
  import_id TEXT NOT NULL,
  org_id TEXT NOT NULL,
  row_number INTEGER NOT NULL,
  status TEXT NOT NULL,
  error_message TEXT NOT NULL DEFAULT '',
  action TEXT NOT NULL DEFAULT 'create',
  existing_equipment_id TEXT,
  existing_item_id TEXT,
  created_at INTEGER NOT NULL DEFAULT (unixepoch()),
  -- raw CSV columns; all TEXT to preserve input before validation
  name TEXT NOT NULL DEFAULT '',
  -- Labels are defined by Gearberg and should match the seeded once
  type_label TEXT NOT NULL DEFAULT '', -- Bulk/Serialized
  tracking_label TEXT NOT NULL DEFAULT '', -- Physical/Virtual
  usage_type_label TEXT NOT NULL DEFAULT '', -- Rental/Resale
  -- User provided fields:
  category_name TEXT NOT NULL DEFAULT '',
  manufacturer_name TEXT NOT NULL DEFAULT '',
  location_name TEXT NOT NULL DEFAULT '',
  purchase_price TEXT NOT NULL DEFAULT '',
  rental_price TEXT NOT NULL DEFAULT '',
  resale_price TEXT NOT NULL DEFAULT '',
  notes TEXT NOT NULL DEFAULT '',
  weight_g TEXT NOT NULL DEFAULT '',
  width_mm TEXT NOT NULL DEFAULT '',
  height_mm TEXT NOT NULL DEFAULT '',
  depth_mm TEXT NOT NULL DEFAULT '',
  voltage_mv TEXT NOT NULL DEFAULT '',
  current_ma TEXT NOT NULL DEFAULT '',
  power_mw TEXT NOT NULL DEFAULT '',
  wire_gauge_mm2_x100 TEXT NOT NULL DEFAULT '',
  quantity TEXT NOT NULL DEFAULT '1',
  equipment_type_label TEXT NOT NULL DEFAULT '',
  -- Units for serialized equipment items
  unit_serial_number TEXT NOT NULL DEFAULT '',
  unit_manufacturer_serial TEXT NOT NULL DEFAULT '',
  unit_purchase_price TEXT NOT NULL DEFAULT '',
  unit_purchased_at TEXT NOT NULL DEFAULT '',
  next_inspection_at TEXT NOT NULL DEFAULT '',
  unit_is_active TEXT NOT NULL DEFAULT '1',
  unit_remark TEXT NOT NULL DEFAULT ''
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS equipment_imports;
-- +goose StatementEnd
