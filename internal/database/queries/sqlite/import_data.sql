-- name: InsertData :one
INSERT INTO import_data (id, session_id, row_number, data)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: GetData :one
SELECT * FROM import_data
WHERE id = ?;

-- name: ListData :many
SELECT * FROM import_data
WHERE session_id = ?
ORDER BY row_number ASC;

-- name: ListDataByAction :many
SELECT * FROM import_data
WHERE session_id = ? AND action = ?
ORDER BY row_number ASC;

-- name: UpdateDataAction :one
UPDATE import_data
SET action = ?, updated_at = unixepoch()
WHERE id = ?
RETURNING *;

-- name: UpdateDataStatus :one
UPDATE import_data
SET status = ?, error_message = ?, updated_at = unixepoch()
WHERE id = ?
RETURNING *;

-- name: DeleteDataBySession :exec
DELETE FROM import_data
WHERE session_id = ?;
