-- name: Create :one
INSERT INTO locations (id, org_id, name)
VALUES (?, ?, ?)
RETURNING id, org_id, name, updated_at, created_at;

-- name: GetByID :one
SELECT id, org_id, name, updated_at, created_at FROM locations WHERE id = ?;

-- name: GetByOrgID :many
SELECT id, org_id, name, updated_at, created_at FROM locations WHERE org_id = ? ORDER BY name ASC;

-- name: Update :one
UPDATE locations SET name = ?, updated_at = unixepoch() WHERE id = ?
RETURNING id, org_id, name, updated_at, created_at;

-- name: Delete :exec
DELETE FROM locations WHERE id = ?;

-- name: CountByOrgID :one
SELECT COUNT(*) FROM locations WHERE org_id = ?;

-- name: GetByName :one
SELECT id, org_id, name, updated_at, created_at FROM locations WHERE org_id = ? AND name = ?;
