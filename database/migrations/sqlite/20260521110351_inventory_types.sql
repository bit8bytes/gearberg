-- +goose Up
-- +goose StatementBegin
CREATE TABLE inventory_types (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL UNIQUE
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS inventory_types;
-- +goose StatementEnd
