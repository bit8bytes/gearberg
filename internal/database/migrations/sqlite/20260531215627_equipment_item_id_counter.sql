-- +goose Up
-- +goose StatementBegin
CREATE TABLE equipment_item_id_counter (
  org_id TEXT PRIMARY KEY REFERENCES orgs(id) ON DELETE CASCADE,
  seq    INTEGER NOT NULL DEFAULT 0
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS equipment_item_id_counter;
-- +goose StatementEnd
