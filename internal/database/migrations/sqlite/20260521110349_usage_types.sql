-- +goose Up
-- +goose StatementBegin
CREATE TABLE usage_types (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL UNIQUE
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS usage_types;
-- +goose StatementEnd
