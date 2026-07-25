-- name: Create :one
INSERT INTO accounts (
    id,
    email,
    email_verified
) VALUES (
    ?, ?, ?
) RETURNING
    id,
    email,
    email_verified,
    enabled,
    created_at;

-- name: GetByAccountID :one
SELECT
    id,
    email,
    email_verified,
    enabled,
    updated_at,
    created_at
FROM accounts
WHERE id = ?
LIMIT 1;

-- name: Exists :one
SELECT EXISTS(
    SELECT 1 FROM accounts WHERE email = ?
) AS account_exists;

-- name: Delete :exec
DELETE FROM accounts
WHERE id = ?;
