-- name: MaxCodeByOrgID :one
SELECT CAST(COALESCE(MAX(code), 0) AS INTEGER) FROM inventory
WHERE org_id = ?;

-- name: List :many
SELECT
    i.id,
    i.org_id,
    i.name,
    i.code,
    i.category_id,
    COALESCE(ec.name, '') AS category_name,
    i.manufacturer_id,
    i.storage_object_id,
    i.type_id,
    i.usage_type_id,
    i.total_stock,
    i.purchase_price,
    i.rental_price,
    i.notes,
    i.updated_at,
    i.created_at,
    COUNT(*) OVER() AS total_records
FROM inventory i
LEFT JOIN equipment_categories ec ON ec.id = i.category_id
WHERE i.org_id = sqlc.arg(org_id)
  AND (sqlc.arg(name_query) = '' OR i.name LIKE '%' || sqlc.arg(name_query) || '%')
  AND (sqlc.arg(category) = '' OR ec.name = sqlc.arg(category))
ORDER BY i.name ASC
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: CountByOrgID :one
SELECT COUNT(*) FROM inventory
WHERE org_id = ?;

-- name: Create :one
INSERT INTO inventory (
    id,
    org_id,
    name,
    category_id,
    manufacturer_id,
    type_id,
    usage_type_id,
    code,
    total_stock,
    purchase_price,
    rental_price,
    notes,
    weight_g,
    width_mm,
    height_mm,
    depth_mm,
    power_mw,
    current_ma
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
    ?,
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
    org_id,
    name,
    category_id,
    manufacturer_id,
    storage_object_id,
    type_id,
    usage_type_id,
    code,
    total_stock,
    purchase_price,
    rental_price,
    notes,
    weight_g,
    width_mm,
    height_mm,
    depth_mm,
    power_mw,
    current_ma,
    created_at;

-- name: CreateUnit :one
INSERT INTO inventory_units (
    id,
    inventory_id,
    status_id,
    unit_number,
    serial_number,
    next_inspection_at
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?,
    ?
) RETURNING
    id,
    inventory_id,
    status_id,
    unit_number,
    serial_number,
    next_inspection_at,
    created_at;


-- name: GetByID :one
SELECT
    id,
    org_id,
    type_id,
    usage_type_id,
    name,
    code,
    category_id,
    manufacturer_id,
    storage_object_id,
    total_stock,
    purchase_price,
    rental_price,
    notes,
    weight_g,
    width_mm,
    height_mm,
    depth_mm,
    power_mw,
    current_ma,
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
    code = ?,
    total_stock = ?,
    purchase_price = ?,
    rental_price = ?,
    notes = ?,
    weight_g = ?,
    width_mm = ?,
    height_mm = ?,
    depth_mm = ?,
    power_mw = ?,
    current_ma = ?,
    updated_at = unixepoch()
WHERE id = ?
RETURNING
    id,
    org_id,
    name,
    category_id,
    manufacturer_id,
    storage_object_id,
    type_id,
    usage_type_id,
    code,
    total_stock,
    purchase_price,
    rental_price,
    notes,
    weight_g,
    width_mm,
    height_mm,
    depth_mm,
    power_mw,
    current_ma,
    updated_at,
    created_at;

-- name: ListUnitsByInventoryID :many
SELECT
    id,
    inventory_id,
    status_id,
    unit_number,
    serial_number,
    notes,
    next_inspection_at,
    updated_at,
    created_at
FROM inventory_units
WHERE inventory_id = ?
ORDER BY unit_number ASC;

-- name: UpdateStorageObject :exec
UPDATE inventory
SET storage_object_id = sqlc.arg(storage_object_id)
WHERE id = sqlc.arg(id);

-- name: Delete :exec
DELETE FROM inventory
WHERE id = ?;

-- name: GetUnit :one
SELECT
    id,
    inventory_id,
    status_id,
    unit_number,
    serial_number,
    notes,
    next_inspection_at,
    updated_at,
    created_at
FROM inventory_units
WHERE id = ?;

-- name: MaxUnitNumber :one
SELECT CAST(COALESCE(MAX(unit_number), 0) AS INTEGER) FROM inventory_units
WHERE inventory_id = ?;

-- name: ListUnitStatuses :many
SELECT id, name FROM unit_statuses ORDER BY id ASC;

-- name: UpdateUnit :exec
UPDATE inventory_units
SET
    status_id = ?,
    serial_number = ?,
    next_inspection_at = ?,
    notes = ?,
    updated_at = unixepoch()
WHERE id = ?;

-- name: DeleteUnit :exec
DELETE FROM inventory_units
WHERE id = ?;

-- name: UpdateTotalStock :exec
UPDATE inventory
SET
    total_stock = total_stock + ?,
    updated_at = unixepoch()
WHERE id = ?;

-- name: ListAllByOrgID :many
SELECT
    i.id,
    i.org_id,
    i.name,
    i.code,
    i.category_id,
    COALESCE(ec.name, '') AS category_name,
    i.manufacturer_id,
    i.storage_object_id,
    i.type_id,
    i.usage_type_id,
    i.total_stock,
    i.purchase_price,
    i.rental_price,
    i.notes,
    i.updated_at,
    i.created_at
FROM inventory i
LEFT JOIN equipment_categories ec ON ec.id = i.category_id
WHERE i.org_id = ?
ORDER BY i.name ASC;
