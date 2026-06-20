-- +goose Up
-- +goose StatementBegin
CREATE TABLE equipment_combination_items (
  id TEXT PRIMARY KEY,
  equipment_id TEXT NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
  member_equipment_id TEXT NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
  quantity INTEGER NOT NULL DEFAULT 1,
  UNIQUE(equipment_id, member_equipment_id)
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS equipment_combination_items;
-- +goose StatementEnd
