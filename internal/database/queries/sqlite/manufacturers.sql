-- name: CountByOrgID :one
SELECT COUNT(*) FROM manufacturers
WHERE org_id = ?;

-- name: Create :one
INSERT INTO manufacturers (
    id,
    org_id,
    name
) VALUES (
    ?,
    ?,
    ?
) RETURNING
    id,
    org_id,
    name,
    created_at;

-- name: GetByOrgID :many
SELECT
    id,
    org_id,
    name,
    updated_at,
    created_at
FROM manufacturers
WHERE org_id = ?;

-- name: GetByID :one
SELECT
    id,
    org_id,
    name,
    updated_at,
    created_at
FROM manufacturers
WHERE id = ?;

-- name: Update :one
UPDATE manufacturers
SET
    name = ?,
    updated_at = unixepoch()
WHERE id = ?
RETURNING
    id,
    org_id,
    name,
    updated_at,
    created_at;

-- name: GetByName :one
SELECT
    id,
    org_id,
    name,
    updated_at,
    created_at
FROM manufacturers
WHERE org_id = ? AND name = ?;

-- name: Delete :exec
DELETE FROM manufacturers
WHERE id = ?;
