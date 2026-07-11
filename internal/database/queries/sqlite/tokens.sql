-- name: GetScopeByName :one
SELECT
    id,
    name
FROM token_scopes
WHERE name = ?
LIMIT 1;

-- name: Insert :one
INSERT INTO tokens (
    account_id,
    token_scope_id,
    hash,
    expires_at
) VALUES (
    ?, ?, ?, ?
) RETURNING
    id,
    account_id,
    token_scope_id,
    expires_at,
    created_at;

-- name: VerifyAndIncrement :one
UPDATE tokens
SET attempts = attempts + 1
WHERE hash = ?
  AND expires_at > unixepoch()
  AND attempts >= 0
  AND attempts <= 5
RETURNING id, account_id, token_scope_id, expires_at, created_at, attempts;

-- name: DeleteForAccountAndScope :exec
DELETE FROM tokens
WHERE account_id = ?
  AND token_scope_id = ?;

-- name: DeleteExpiredTokens :exec
DELETE FROM tokens
WHERE expires_at <= unixepoch();
