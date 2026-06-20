-- name: List :many
SELECT
    e.id,
    e.org_id,
    e.name,
    e.category_id,
    COALESCE(ec.name, '') AS category_name,
    e.manufacturer_id,
    e.location_id,
    COALESCE(wl.name, '') AS location_name,
    e.storage_object_id,
    e.tracking_type_id,
    COALESCE(tt.name, '') AS tracking_type_name,
    e.usage_type_id,
    e.has_content,
    CASE
        WHEN tt.name = 'bulk'       THEN COALESCE((SELECT SUM(ei.quantity) FROM equipment_items ei WHERE ei.equipment_id = e.id AND ei.parent_equipment_item_id IS NULL), 0)
        WHEN tt.name = 'serialized' THEN (SELECT COUNT(*) FROM equipment_items ei WHERE ei.equipment_id = e.id)
        ELSE 0
    END AS total_stock,
    e.rental_price,
    e.resale_price,
    e.notes,
    e.updated_at,
    e.created_at,
    COUNT(*) OVER() AS total_records
FROM equipment e
LEFT JOIN equipment_categories ec ON ec.id = e.category_id
LEFT JOIN warehouse_locations wl ON wl.id = e.location_id
LEFT JOIN tracking_types tt ON tt.id = e.tracking_type_id
WHERE e.org_id = sqlc.arg(org_id)
  AND (sqlc.arg(name_query) = '' OR e.name LIKE '%' || sqlc.arg(name_query) || '%')
  AND (sqlc.arg(category) = '' OR ec.name = sqlc.arg(category))
ORDER BY e.name ASC
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: CountByOrgID :one
SELECT COUNT(*) FROM equipment
WHERE org_id = ?;

-- name: Create :one
INSERT INTO equipment (
    id,
    org_id,
    tracking_type_id,
    category_id,
    manufacturer_id,
    usage_type_id,
    location_id,
    name,
    has_content,
    rental_price,
    resale_price,
    notes,
    weight_g,
    width_mm,
    height_mm,
    depth_mm,
    voltage_v,
    current_ma,
    power_mw
) VALUES (
    sqlc.arg(id),
    sqlc.arg(org_id),
    sqlc.arg(tracking_type_id),
    sqlc.arg(category_id),
    sqlc.arg(manufacturer_id),
    sqlc.arg(usage_type_id),
    sqlc.arg(location_id),
    sqlc.arg(name),
    sqlc.arg(has_content),
    sqlc.arg(rental_price),
    sqlc.arg(resale_price),
    sqlc.arg(notes),
    sqlc.arg(weight_g),
    sqlc.arg(width_mm),
    sqlc.arg(height_mm),
    sqlc.arg(depth_mm),
    sqlc.arg(voltage_v),
    sqlc.arg(current_ma),
    sqlc.arg(power_mw)
) RETURNING
    id,
    org_id,
    tracking_type_id,
    category_id,
    manufacturer_id,
    usage_type_id,
    location_id,
    storage_object_id,
    name,
    has_content,
    rental_price,
    resale_price,
    notes,
    weight_g,
    width_mm,
    height_mm,
    depth_mm,
    voltage_v,
    current_ma,
    power_mw,
    created_at;

-- name: GetByID :one
SELECT
    e.id,
    e.org_id,
    e.tracking_type_id,
    e.category_id,
    e.manufacturer_id,
    e.usage_type_id,
    e.location_id,
    COALESCE(wl.name, '') AS location_name,
    e.storage_object_id,
    e.name,
    e.has_content,
    CASE
        WHEN tt.name = 'bulk'       THEN COALESCE((SELECT SUM(ei.quantity) FROM equipment_items ei WHERE ei.equipment_id = e.id AND ei.parent_equipment_item_id IS NULL), 0)
        WHEN tt.name = 'serialized' THEN (SELECT COUNT(*) FROM equipment_items ei WHERE ei.equipment_id = e.id)
        ELSE 0
    END AS total_stock,
    (SELECT COUNT(*) FROM equipment_combination_items WHERE equipment_id = e.id) AS content_count,
    e.rental_price,
    e.resale_price,
    e.notes,
    e.weight_g,
    e.width_mm,
    e.height_mm,
    e.depth_mm,
    e.voltage_v,
    e.current_ma,
    e.power_mw,
    e.updated_at,
    e.created_at
FROM equipment e
LEFT JOIN warehouse_locations wl ON wl.id = e.location_id
LEFT JOIN tracking_types tt ON tt.id = e.tracking_type_id
WHERE e.id = ?;

-- name: UpdateDetails :exec
UPDATE equipment
SET
    name = sqlc.arg(name),
    category_id = sqlc.arg(category_id),
    manufacturer_id = sqlc.arg(manufacturer_id),
    location_id = sqlc.arg(location_id),
    notes = sqlc.arg(notes),
    updated_at = unixepoch()
WHERE id = sqlc.arg(id);

-- name: UpdatePricing :exec
UPDATE equipment
SET
    rental_price = sqlc.arg(rental_price),
    resale_price = sqlc.arg(resale_price),
    updated_at = unixepoch()
WHERE id = sqlc.arg(id);

-- name: UpdateProperties :exec
UPDATE equipment
SET
    weight_g = sqlc.arg(weight_g),
    width_mm = sqlc.arg(width_mm),
    height_mm = sqlc.arg(height_mm),
    depth_mm = sqlc.arg(depth_mm),
    voltage_v = sqlc.arg(voltage_v),
    current_ma = sqlc.arg(current_ma),
    power_mw = sqlc.arg(power_mw),
    updated_at = unixepoch()
WHERE id = sqlc.arg(id);

-- name: UpdateStorageObject :exec
UPDATE equipment
SET storage_object_id = sqlc.arg(storage_object_id)
WHERE id = sqlc.arg(id);

-- name: Delete :exec
DELETE FROM equipment
WHERE id = ?;

-- name: ListAllByOrgID :many
SELECT
    e.id,
    e.org_id,
    e.name,
    e.category_id,
    COALESCE(ec.name, '') AS category_name,
    e.manufacturer_id,
    e.storage_object_id,
    e.tracking_type_id,
    e.usage_type_id,
    e.has_content,
    CASE
        WHEN tt.name = 'bulk'       THEN COALESCE((SELECT SUM(ei.quantity) FROM equipment_items ei WHERE ei.equipment_id = e.id AND ei.parent_equipment_item_id IS NULL), 0)
        WHEN tt.name = 'serialized' THEN (SELECT COUNT(*) FROM equipment_items ei WHERE ei.equipment_id = e.id)
        ELSE 0
    END AS total_stock,
    e.rental_price,
    e.resale_price,
    e.notes,
    e.updated_at,
    e.created_at
FROM equipment e
LEFT JOIN equipment_categories ec ON ec.id = e.category_id
LEFT JOIN tracking_types tt ON tt.id = e.tracking_type_id
WHERE e.org_id = ?
ORDER BY e.name ASC;
