-- name: InsertSession :one
INSERT INTO import_sessions (id, org_id, format, status, target_entity)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetSession :one
SELECT * FROM import_sessions
WHERE id = ?;

-- name: UpdateSessionStatus :one
UPDATE import_sessions
SET status = ?, updated_at = unixepoch()
WHERE id = ?
RETURNING *;

-- name: GetStagedSession :one
SELECT * FROM import_sessions
WHERE org_id = ? AND status = ?
ORDER BY created_at DESC
LIMIT 1;

-- name: DeleteSession :exec
DELETE FROM import_sessions
WHERE id = ?;
