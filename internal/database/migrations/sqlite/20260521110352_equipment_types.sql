-- +goose Up
-- +goose StatementBegin
CREATE TABLE equipment_types (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL UNIQUE
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS equipment_types;
-- +goose StatementEnd
