-- name: Create :one
INSERT INTO warehouse_locations (id, parent_warehouse_location_id, org_id, name)
VALUES (?, ?, ?, ?)
RETURNING id, parent_warehouse_location_id, org_id, name;

-- name: GetByID :one
SELECT id, parent_warehouse_location_id, org_id, name
FROM warehouse_locations
WHERE id = ?;

-- name: GetByOrgID :many
SELECT id, parent_warehouse_location_id, org_id, name
FROM warehouse_locations
WHERE org_id = ?
ORDER BY name ASC;

-- name: Update :one
UPDATE warehouse_locations
SET name = ?
WHERE id = ?
RETURNING id, parent_warehouse_location_id, org_id, name;

-- name: Delete :exec
DELETE FROM warehouse_locations
WHERE id = ?;

-- name: CountByOrgID :one
SELECT COUNT(*) FROM warehouse_locations
WHERE org_id = ?;

-- name: GetByName :one
SELECT id, parent_warehouse_location_id, org_id, name
FROM warehouse_locations
WHERE org_id = ? AND name = ?;
