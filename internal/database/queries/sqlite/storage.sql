-- name: Create :one
INSERT INTO storage_objects (
    id,
    org_id,
    key,
    backend,
    filename,
    content_type,
    size,
    encryption_key_id,
    created_at
) VALUES (
    sqlc.arg(id),
    sqlc.arg(org_id),
    sqlc.arg(key),
    sqlc.arg(backend),
    sqlc.arg(filename),
    sqlc.arg(content_type),
    sqlc.arg(size),
    sqlc.arg(encryption_key_id),
    unixepoch()
)
ON CONFLICT(key) DO UPDATE SET
    filename         = excluded.filename,
    content_type     = excluded.content_type,
    size             = excluded.size,
    encryption_key_id = excluded.encryption_key_id
RETURNING id, org_id, key, backend, filename, content_type, size, encryption_key_id, created_at;

-- name: Get :one
SELECT id, org_id, key, backend, filename, content_type, size, encryption_key_id, created_at
FROM storage_objects
WHERE key = sqlc.arg(key);

-- name: GetByID :one
SELECT id, org_id, key, backend, filename, content_type, size, encryption_key_id, created_at
FROM storage_objects
WHERE id = sqlc.arg(id);

-- name: Delete :exec
DELETE FROM storage_objects
WHERE id = sqlc.arg(id);

-- name: GetStorageUsedByOrg :one
SELECT CAST(COALESCE(SUM(so.size), 0) AS INTEGER) AS bytes_used
FROM storage_objects so
INNER JOIN equipment e ON e.storage_object_id = so.id
WHERE e.org_id = sqlc.arg(org_id);
