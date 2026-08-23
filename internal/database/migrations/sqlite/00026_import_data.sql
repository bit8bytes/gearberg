-- +goose Up
-- +goose StatementBegin
CREATE TABLE import_data (
  id            TEXT    NOT NULL PRIMARY KEY,
  session_id    TEXT    NOT NULL REFERENCES import_sessions(id) ON DELETE CASCADE,
  row_number    INTEGER NOT NULL CHECK (row_number > 0),
  data          TEXT    NOT NULL DEFAULT '{}' CHECK (json_valid(data)),
  status        TEXT    NOT NULL DEFAULT 'new'
                        CHECK (status IN ('new', 'valid', 'error')),
  error_message TEXT    NOT NULL DEFAULT '',
  action        TEXT    NOT NULL DEFAULT 'pending'
                        CHECK (action IN ('pending', 'create', 'skip')),
  updated_at    INTEGER NOT NULL DEFAULT (unixepoch()),
  created_at    INTEGER NOT NULL DEFAULT (unixepoch()),
  UNIQUE (session_id, row_number)
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS import_data;
-- +goose StatementEnd
