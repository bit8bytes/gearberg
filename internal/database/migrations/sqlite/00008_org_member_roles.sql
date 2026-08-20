-- +goose Up
-- +goose StatementBegin

-- Table: org_members
-- Associates accounts with orgss and their roles within those orgss
CREATE TABLE IF NOT EXISTS org_members (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    account_id TEXT NOT NULL,
    org_id TEXT NOT NULL,
    role_id INTEGER NOT NULL,
    created_at INTEGER NOT NULL DEFAULT (unixepoch()),
    updated_at INTEGER NOT NULL DEFAULT (unixepoch()),
    FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE,
    FOREIGN KEY (org_id) REFERENCES orgs(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES org_roles(id) ON DELETE RESTRICT,
    UNIQUE(account_id, org_id)
) STRICT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS org_members;
-- +goose StatementEnd
