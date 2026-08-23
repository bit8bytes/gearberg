-- +goose Up
-- +goose StatementBegin
CREATE TABLE import_sessions (
  id            TEXT    NOT NULL PRIMARY KEY,
  org_id        TEXT    NOT NULL REFERENCES orgs(id) ON DELETE CASCADE,
  format        TEXT    NOT NULL, -- csv | json | excel
  status        TEXT    NOT NULL, -- uploading | mapping | staged | committed
  target_entity TEXT    NOT NULL, -- equipment | ...
  created_at    INTEGER NOT NULL DEFAULT (unixepoch())
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS import_sessions;
-- +goose StatementEnd
