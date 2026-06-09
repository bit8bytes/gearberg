-- +goose Up
-- +goose StatementBegin
CREATE TABLE serialized_units (
  id TEXT PRIMARY KEY,
  inventory_id TEXT NOT NULL REFERENCES inventory(id) ON DELETE CASCADE,
  status_id INTEGER NOT NULL REFERENCES unit_statuses(id) ON DELETE RESTRICT,
  unit_number INTEGER NOT NULL,
  serial_number TEXT,
  notes TEXT,
  next_inspection_at INTEGER,
  updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
  created_at INTEGER NOT NULL DEFAULT (unixepoch()),
  UNIQUE(inventory_id, unit_number)
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS serialized_units;
-- +goose StatementEnd
