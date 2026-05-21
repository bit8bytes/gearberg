-- name: Create :one
INSERT INTO company_settings (
    id,
    company_id,
    currency,
    vat_rate,
    timezone
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?
) RETURNING
    id,
    company_id,
    currency,
    vat_rate,
    timezone,
    created_at;

-- name: GetByID :one
SELECT
    id,
    company_id,
    currency,
    vat_rate,
    timezone,
    updated_at,
    created_at
FROM company_settings
WHERE id = ?;

-- name: GetByCompanyID :one
SELECT
    id,
    company_id,
    currency,
    vat_rate,
    timezone,
    updated_at,
    created_at
FROM company_settings
WHERE company_id = ?;

-- name: Update :one
UPDATE company_settings
SET
    currency = ?,
    vat_rate = ?,
    timezone = ?,
    updated_at = unixepoch()
WHERE id = ?
RETURNING
    id,
    company_id,
    currency,
    vat_rate,
    timezone,
    updated_at,
    created_at;

-- name: Delete :exec
DELETE FROM company_settings
WHERE id = ?;
