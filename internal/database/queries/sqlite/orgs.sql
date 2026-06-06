-- name: Create :one
INSERT INTO orgs (
    id,
    name
) VALUES (
    ?,
    ?
) RETURNING 
    id,
    name,
    created_at;

-- name: GetAll :many
SELECT
    id,
    name,
    updated_at,
    created_at
FROM orgs;

-- name: GetByID :one
SELECT
    id,
    name,
    updated_at,
    created_at
FROM orgs
WHERE id = ?;
-- name: UpdateByID :one
UPDATE orgs
SET
    name = ?,
    updated_at = unixepoch()
WHERE id = ?
RETURNING
    id,
    name,
    updated_at,
    created_at;

-- name: DeleteByID :exec
DELETE FROM orgs
WHERE id = ?;

-- name: CountOrgs :one
SELECT COUNT(*) FROM orgs;
