-- +goose Up
-- +goose StatementBegin
CREATE TABLE import_data (
  id            TEXT    NOT NULL PRIMARY KEY,
  session_id    TEXT    NOT NULL REFERENCES import_sessions(id) ON DELETE CASCADE,
  row_number    INTEGER NOT NULL CHECK (row_number > 0),
  data          TEXT    NOT NULL DEFAULT '{}', -- JSON blob of raw source columns
  status        TEXT    NOT NULL DEFAULT 'new', -- new | error | needs_review
  error_message TEXT    NOT NULL DEFAULT '',
  action        TEXT    NOT NULL DEFAULT 'pending', -- create | skip | pending
  UNIQUE (session_id, row_number)
) STRICT;

CREATE INDEX import_data_session_id_idx ON import_data (session_id);
CREATE INDEX import_data_action_idx ON import_data (session_id, action);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS import_data;
-- +goose StatementEnd
