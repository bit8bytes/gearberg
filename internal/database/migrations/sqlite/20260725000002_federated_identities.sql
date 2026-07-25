-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS federated_identities (
    account_id TEXT NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    provider_id INTEGER NOT NULL REFERENCES provider_types(id) ON DELETE RESTRICT,
    provider_subject TEXT NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
    PRIMARY KEY (account_id, provider_id),
    UNIQUE (provider_id, provider_subject)
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS federated_identities;
-- +goose StatementEnd
