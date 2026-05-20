-- name: Create :one
INSERT INTO companies (
    id,
    name
) VALUES (
    ?,
    ?
) RETURNING 
    id,
    name,
    created_at;

-- name: GetAll :many
SELECT
    id,
    name,
    updated_at,
    created_at
FROM companies;

-- name: GetByID :one
SELECT
    id,
    name,
    updated_at,
    created_at
FROM companies
WHERE id = ?;
-- name: UpdateByID :one
UPDATE companies
SET
    name = ?,
    updated_at = unixepoch()
WHERE id = ?
RETURNING
    id,
    name,
    updated_at,
    created_at;

-- name: DeleteByID :exec
DELETE FROM companies
WHERE id = ?;

-- name: CountCompanies :one
SELECT COUNT(*) FROM companies;
