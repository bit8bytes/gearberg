-- +goose Up
-- +goose StatementBegin
CREATE TABLE unit_inspections (
  id TEXT PRIMARY KEY,
  unit_id TEXT NOT NULL REFERENCES serialized_units(id) ON DELETE CASCADE,
  inspected_at INTEGER NOT NULL,
  passed INTEGER NOT NULL DEFAULT 1,
  notes TEXT,
  created_at INTEGER NOT NULL DEFAULT (unixepoch())
) STRICT;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_unit_inspections_unit_id ON unit_inspections(unit_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS unit_inspections;
-- +goose StatementEnd
