-- name: Create :exec
INSERT INTO credentials (
    account_id,
    type_id,
    secret_data
) VALUES (
    ?, ?, ?
);

-- name: GetByEmail :one
SELECT
    c.secret_data,
    a.id AS account_id
FROM credentials c
JOIN accounts a ON a.id = c.account_id
WHERE a.email = ?
AND c.type_id = ?
LIMIT 1;

-- name: GetByAccountID :one
SELECT
    id,
    account_id,
    type_id,
    secret_data,
    updated_at,
    created_at
FROM credentials
WHERE account_id = ?
AND type_id = ?
LIMIT 1;

-- name: UpdateSecret :exec
UPDATE credentials
SET
    secret_data = ?,
    updated_at = unixepoch()
WHERE account_id = ?
AND type_id = ?;
