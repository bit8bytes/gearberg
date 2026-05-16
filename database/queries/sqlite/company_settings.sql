-- name: CreateCompanySettings :one
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

-- name: GetCompanySettingsByID :one
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

-- name: GetCompanySettingsByCompanyID :one
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

-- name: UpdateCompanySettings :one
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

-- name: DeleteCompanySettings :exec
DELETE FROM company_settings
WHERE id = ?;
