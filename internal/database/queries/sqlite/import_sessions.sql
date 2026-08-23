-- name: InsertSession :one
INSERT INTO import_sessions (id, org_id, format, status, target_entity)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: GetSession :one
SELECT * FROM import_sessions
WHERE id = ?;

-- name: UpdateSessionStatus :one
UPDATE import_sessions
SET status = ?
WHERE id = ?
RETURNING *;

-- name: DeleteSession :exec
DELETE FROM import_sessions
WHERE id = ?;
