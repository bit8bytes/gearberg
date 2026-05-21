-- name: CountByCompanyID :one
SELECT COUNT(*) FROM manufacturers
WHERE company_id = ?;

-- name: Create :one
INSERT INTO manufacturers (
    id,
    company_id,
    name
) VALUES (
    ?,
    ?,
    ?
) RETURNING
    id,
    company_id,
    name,
    created_at;

-- name: GetByCompanyID :many
SELECT
    id,
    company_id,
    name,
    updated_at,
    created_at
FROM manufacturers
WHERE company_id = ?;

-- name: GetByID :one
SELECT
    id,
    company_id,
    name,
    updated_at,
    created_at
FROM manufacturers
WHERE id = ?;

-- name: Update :one
UPDATE manufacturers
SET
    name = ?,
    updated_at = unixepoch()
WHERE id = ?
RETURNING
    id,
    company_id,
    name,
    updated_at,
    created_at;

-- name: Delete :exec
DELETE FROM manufacturers
WHERE id = ?;
