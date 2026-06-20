-- name: CountByOrgID :one
SELECT COUNT(*) FROM equipment_manufacturers
WHERE org_id = ?;

-- name: Create :one
INSERT INTO equipment_manufacturers (
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
FROM equipment_manufacturers
WHERE org_id = ?;

-- name: GetByID :one
SELECT
    id,
    org_id,
    name,
    updated_at,
    created_at
FROM equipment_manufacturers
WHERE id = ?;

-- name: Update :one
UPDATE equipment_manufacturers
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
FROM equipment_manufacturers
WHERE org_id = ? AND name = ?;

-- name: Delete :exec
DELETE FROM equipment_manufacturers
WHERE id = ?;
