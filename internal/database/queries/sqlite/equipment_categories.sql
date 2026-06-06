-- name: Create :one
INSERT INTO equipment_categories (
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

-- name: GetByID :one
SELECT
    id,
    org_id,
    name,
    updated_at,
    created_at
FROM equipment_categories
WHERE id = ?;

-- name: GetAll :many
SELECT
    id,
    org_id,
    name,
    updated_at,
    created_at
FROM equipment_categories;

-- name: GetByOrgID :many
SELECT
    id,
    org_id,
    name,
    updated_at,
    created_at
FROM equipment_categories
WHERE org_id = ?;

-- name: Update :one
UPDATE equipment_categories
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

-- name: Delete :exec
DELETE FROM equipment_categories
WHERE id = ?;

-- name: CountByOrgID :one
SELECT COUNT(*) FROM equipment_categories
WHERE org_id = ?;
