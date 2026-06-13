-- name: List :many
SELECT
    i.id,
    i.org_id,
    i.name,
    i.code,
    i.category_id,
    COALESCE(ec.name, '') AS category_name,
    COALESCE(ec.color, '') AS category_color,
    i.manufacturer_id,
    i.storage_object_id,
    i.type_id,
    i.usage_type_id,
    COALESCE(bs.quantity, (SELECT COUNT(*) FROM serialized_units su WHERE su.inventory_id = i.id)) AS total_stock,
    i.purchase_price,
    i.rental_price,
    i.notes,
    i.updated_at,
    i.created_at,
    COUNT(*) OVER() AS total_records
FROM inventory i
LEFT JOIN equipment_categories ec ON ec.id = i.category_id
LEFT JOIN bulk_stock bs ON bs.inventory_id = i.id
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
    sqlc.arg(id),
    sqlc.arg(org_id),
    sqlc.arg(name),
    sqlc.arg(category_id),
    sqlc.arg(manufacturer_id),
    sqlc.arg(type_id),
    sqlc.arg(usage_type_id),
    (SELECT COALESCE(MAX(code), 0) + 1 FROM inventory WHERE org_id = sqlc.arg(org_id)),
    sqlc.arg(purchase_price),
    sqlc.arg(rental_price),
    sqlc.arg(notes),
    sqlc.arg(weight_g),
    sqlc.arg(width_mm),
    sqlc.arg(height_mm),
    sqlc.arg(depth_mm),
    sqlc.arg(power_mw),
    sqlc.arg(current_ma)
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

-- name: GetByID :one
SELECT
    i.id,
    i.org_id,
    i.type_id,
    i.usage_type_id,
    i.name,
    i.code,
    i.category_id,
    i.manufacturer_id,
    i.storage_object_id,
    COALESCE(bs.quantity, (SELECT COUNT(*) FROM serialized_units su WHERE su.inventory_id = i.id)) AS total_stock,
    i.purchase_price,
    i.rental_price,
    i.notes,
    i.weight_g,
    i.width_mm,
    i.height_mm,
    i.depth_mm,
    i.power_mw,
    i.current_ma,
    i.inspection_interval_days,
    i.updated_at,
    i.created_at
FROM inventory i
LEFT JOIN bulk_stock bs ON bs.inventory_id = i.id
WHERE i.id = ?;

-- name: UpdateDetails :exec
UPDATE inventory
SET
    name = sqlc.arg(name),
    category_id = sqlc.arg(category_id),
    manufacturer_id = sqlc.arg(manufacturer_id),
    notes = sqlc.arg(notes),
    updated_at = unixepoch()
WHERE id = sqlc.arg(id);

-- name: UpdatePricing :exec
UPDATE inventory
SET
    purchase_price = sqlc.arg(purchase_price),
    rental_price = sqlc.arg(rental_price),
    updated_at = unixepoch()
WHERE id = sqlc.arg(id);

-- name: UpdateProperties :exec
UPDATE inventory
SET
    weight_g = sqlc.arg(weight_g),
    width_mm = sqlc.arg(width_mm),
    height_mm = sqlc.arg(height_mm),
    depth_mm = sqlc.arg(depth_mm),
    power_mw = sqlc.arg(power_mw),
    current_ma = sqlc.arg(current_ma),
    updated_at = unixepoch()
WHERE id = sqlc.arg(id);

-- name: UpdateInspection :exec
UPDATE inventory
SET
    inspection_interval_days = sqlc.arg(inspection_interval_days),
    updated_at = unixepoch()
WHERE id = sqlc.arg(id);

-- name: UpdateStorageObject :exec
UPDATE inventory
SET storage_object_id = sqlc.arg(storage_object_id)
WHERE id = sqlc.arg(id);

-- name: Delete :exec
DELETE FROM inventory
WHERE id = ?;

-- name: ListUnitStatuses :many
SELECT id, name FROM unit_statuses ORDER BY id ASC;

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
    COALESCE(bs.quantity, (SELECT COUNT(*) FROM serialized_units su WHERE su.inventory_id = i.id)) AS total_stock,
    i.purchase_price,
    i.rental_price,
    i.notes,
    i.updated_at,
    i.created_at
FROM inventory i
LEFT JOIN equipment_categories ec ON ec.id = i.category_id
LEFT JOIN bulk_stock bs ON bs.inventory_id = i.id
WHERE i.org_id = ?
ORDER BY i.name ASC;
