-- +goose Up
-- +goose StatementBegin
CREATE TABLE bulk_stock (
  inventory_id TEXT PRIMARY KEY REFERENCES inventory(id) ON DELETE CASCADE,
  quantity INTEGER NOT NULL DEFAULT 0,
  updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
  created_at INTEGER NOT NULL DEFAULT (unixepoch())
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS bulk_stock;
-- +goose StatementEnd
