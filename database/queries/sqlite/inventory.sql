-- name: CountByCompanyID :one
SELECT COUNT(*) FROM inventory
WHERE company_id = ?;

-- name: Create :one
INSERT INTO inventory (
    id,
    company_id,
    name,
    category_id,
    manufacturer_id,
    total_stock,
    purchase_price,
    rental_price,
    notes
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
) RETURNING
    id,
    company_id,
    name,
    category_id,
    manufacturer_id,
    storage_object_id,
    total_stock,
    purchase_price,
    rental_price,
    notes,
    created_at;

-- name: GetByCompanyID :many
SELECT
    i.id,
    i.company_id,
    i.name,
    i.category_id,
    COALESCE(ec.name, '') AS category_name,
    i.manufacturer_id,
    i.storage_object_id,
    i.total_stock,
    i.purchase_price,
    i.rental_price,
    i.notes,
    i.updated_at,
    i.created_at
FROM inventory i
LEFT JOIN equipment_categories ec ON ec.id = i.category_id
WHERE i.company_id = ?;

-- name: GetByID :one
SELECT
    id,
    company_id,
    name,
    category_id,
    manufacturer_id,
    storage_object_id,
    total_stock,
    purchase_price,
    rental_price,
    notes,
    updated_at,
    created_at
FROM inventory
WHERE id = ?;

-- name: Update :one
UPDATE inventory
SET
    name = ?,
    category_id = ?,
    manufacturer_id = ?,
    total_stock = ?,
    purchase_price = ?,
    rental_price = ?,
    notes = ?,
    updated_at = unixepoch()
WHERE id = ?
RETURNING
    id,
    company_id,
    name,
    category_id,
    manufacturer_id,
    storage_object_id,
    total_stock,
    purchase_price,
    rental_price,
    notes,
    updated_at,
    created_at;

-- name: UpdateStorageObject :exec
UPDATE inventory
SET storage_object_id = sqlc.arg(storage_object_id)
WHERE id = sqlc.arg(id);

-- name: Delete :exec
DELETE FROM inventory
WHERE id = ?;
