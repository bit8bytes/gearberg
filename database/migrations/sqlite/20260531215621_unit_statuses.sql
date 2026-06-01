-- +goose Up
-- +goose StatementBegin
CREATE TABLE unit_statuses (
  id INTEGER PRIMARY KEY,
  name TEXT NOT NULL UNIQUE
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS unit_statuses;
-- +goose StatementEnd
