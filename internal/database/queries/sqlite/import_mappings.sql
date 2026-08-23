-- name: InsertMapping :one
INSERT INTO import_mappings (id, session_id, source_col, target_field)
VALUES (?, ?, ?, ?)
RETURNING *;

-- name: ListMappings :many
SELECT * FROM import_mappings
WHERE session_id = ?
ORDER BY source_col ASC;

-- name: DeleteMappingsBySession :exec
DELETE FROM import_mappings
WHERE session_id = ?;
