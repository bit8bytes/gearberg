-- name: Create :one
INSERT INTO org_settings (
    id,
    org_id,
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
    org_id,
    currency,
    vat_rate,
    timezone,
    created_at;

-- name: GetByID :one
SELECT
    id,
    org_id,
    currency,
    vat_rate,
    timezone,
    updated_at,
    created_at
FROM org_settings
WHERE id = ?;

-- name: GetByOrgID :one
SELECT
    id,
    org_id,
    currency,
    vat_rate,
    timezone,
    updated_at,
    created_at
FROM org_settings
WHERE org_id = ?;

-- name: Update :one
UPDATE org_settings
SET
    currency = ?,
    vat_rate = ?,
    timezone = ?,
    updated_at = unixepoch()
WHERE id = ?
RETURNING
    id,
    org_id,
    currency,
    vat_rate,
    timezone,
    updated_at,
    created_at;

-- name: Delete :exec
DELETE FROM org_settings
WHERE id = ?;
