-- +goose Up
-- +goose StatementBegin
CREATE TABLE import_mappings (
  id           TEXT NOT NULL PRIMARY KEY,
  session_id   TEXT NOT NULL REFERENCES import_sessions(id) ON DELETE CASCADE,
  source_col   TEXT NOT NULL, -- user column name e.g. "Bezeichnung"
  target_field TEXT NOT NULL, -- internal field name e.g. "equipment.name"
  UNIQUE (session_id, source_col)
) STRICT;

CREATE INDEX import_mappings_session_id_idx ON import_mappings (session_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS import_mappings;
-- +goose StatementEnd
