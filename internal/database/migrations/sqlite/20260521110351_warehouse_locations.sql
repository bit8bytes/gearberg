-- +goose Up
-- +goose StatementBegin
CREATE TABLE warehouse_locations (
  id TEXT PRIMARY KEY,
  parent_warehouse_location_id TEXT REFERENCES warehouse_locations(id) ON DELETE SET NULL,
  org_id TEXT NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
  name TEXT NOT NULL
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS warehouse_locations;
-- +goose StatementEnd
