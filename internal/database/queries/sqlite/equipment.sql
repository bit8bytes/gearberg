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
    e.equipment_type_id,
    COALESCE(et.name, '') AS equipment_type_name,
    e.tracking_type_id,
    COALESCE(tt.name, '') AS tracking_type_name,
    e.usage_type_id,
    CASE
        WHEN tt.name = 'bulk'       THEN COALESCE((SELECT SUM(ebi.quantity) FROM equipment_bulk_items ebi WHERE ebi.equipment_id = e.id), 0)
        WHEN tt.name = 'serialized' THEN (SELECT COUNT(*) FROM equipment_serialized_items esi WHERE esi.equipment_id = e.id)
        ELSE 0
    END AS total_stock,
    e.is_archived,
    e.rental_price,
    e.resale_price,
    e.notes,
    e.weight_g,
    e.width_mm,
    e.height_mm,
    e.depth_mm,
    e.voltage_mv,
    e.current_ma,
    e.power_mw,
    e.wire_gauge_mm2_x100,
    e.updated_at,
    e.created_at,
    COUNT(*) OVER() AS total_records
FROM equipment e
LEFT JOIN equipment_categories ec ON ec.id = e.category_id
LEFT JOIN warehouse_locations wl ON wl.id = e.location_id
LEFT JOIN equipment_types et ON et.id = e.equipment_type_id
LEFT JOIN tracking_types tt ON tt.id = e.tracking_type_id
WHERE e.org_id = sqlc.arg(org_id)
  AND (sqlc.arg(name_query) = '' OR e.name LIKE '%' || sqlc.arg(name_query) || '%' OR EXISTS (SELECT 1 FROM equipment_serialized_items esi WHERE esi.equipment_id = e.id AND esi.serial_number LIKE '%' || sqlc.arg(name_query) || '%'))
  AND (sqlc.arg(category) = '' OR ec.name = sqlc.arg(category))
  AND (sqlc.arg(is_archived) = -1 OR e.is_archived = sqlc.arg(is_archived))
  AND (sqlc.arg(inspection_filter) = ''
    OR (sqlc.arg(inspection_filter) = 'overdue'  AND EXISTS (SELECT 1 FROM equipment_serialized_items esi WHERE esi.equipment_id = e.id AND esi.next_inspection_at IS NOT NULL AND esi.next_inspection_at < unixepoch()))
    OR (sqlc.arg(inspection_filter) = 'due-30d'  AND EXISTS (SELECT 1 FROM equipment_serialized_items esi WHERE esi.equipment_id = e.id AND esi.next_inspection_at IS NOT NULL AND esi.next_inspection_at >= unixepoch() AND esi.next_inspection_at <= unixepoch() + 30 * 86400)))
ORDER BY category_name ASC, e.name ASC
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: ListBySerialNumber :many
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
    e.equipment_type_id,
    COALESCE(et.name, '') AS equipment_type_name,
    e.tracking_type_id,
    COALESCE(tt.name, '') AS tracking_type_name,
    e.usage_type_id,
    CASE
        WHEN tt.name = 'bulk'       THEN COALESCE((SELECT SUM(ebi.quantity) FROM equipment_bulk_items ebi WHERE ebi.equipment_id = e.id), 0)
        WHEN tt.name = 'serialized' THEN (SELECT COUNT(*) FROM equipment_serialized_items esi WHERE esi.equipment_id = e.id)
        ELSE 0
    END AS total_stock,
    e.is_archived,
    e.rental_price,
    e.resale_price,
    e.notes,
    e.updated_at,
    e.created_at,
    COUNT(*) OVER() AS total_records
FROM equipment e
LEFT JOIN equipment_categories ec ON ec.id = e.category_id
LEFT JOIN warehouse_locations wl ON wl.id = e.location_id
LEFT JOIN equipment_types et ON et.id = e.equipment_type_id
LEFT JOIN tracking_types tt ON tt.id = e.tracking_type_id
WHERE e.org_id = sqlc.arg(org_id)
  AND (sqlc.arg(name_query) = '' OR e.name LIKE '%' || sqlc.arg(name_query) || '%' OR EXISTS (SELECT 1 FROM equipment_serialized_items esi WHERE esi.equipment_id = e.id AND esi.serial_number LIKE '%' || sqlc.arg(name_query) || '%'))
  AND (sqlc.arg(category) = '' OR ec.name = sqlc.arg(category))
  AND e.is_archived = sqlc.arg(is_archived)
ORDER BY (SELECT MIN(esi.serial_number) FROM equipment_serialized_items esi WHERE esi.equipment_id = e.id) ASC
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: CountByOrgID :one
SELECT COUNT(*) FROM equipment
WHERE org_id = ?;

-- name: Create :one
INSERT INTO equipment (
    id,
    org_id,
    equipment_type_id,
    tracking_type_id,
    category_id,
    manufacturer_id,
    usage_type_id,
    location_id,
    name,
    is_archived,
    rental_price,
    resale_price,
    notes,
    weight_g,
    width_mm,
    height_mm,
    depth_mm,
    voltage_mv,
    current_ma,
    power_mw,
    wire_gauge_mm2_x100
) VALUES (
    sqlc.arg(id),
    sqlc.arg(org_id),
    sqlc.arg(equipment_type_id),
    sqlc.arg(tracking_type_id),
    sqlc.arg(category_id),
    sqlc.arg(manufacturer_id),
    sqlc.arg(usage_type_id),
    sqlc.arg(location_id),
    sqlc.arg(name),
    sqlc.arg(is_archived),
    sqlc.arg(rental_price),
    sqlc.arg(resale_price),
    sqlc.arg(notes),
    sqlc.arg(weight_g),
    sqlc.arg(width_mm),
    sqlc.arg(height_mm),
    sqlc.arg(depth_mm),
    sqlc.arg(voltage_mv),
    sqlc.arg(current_ma),
    sqlc.arg(power_mw),
    sqlc.arg(wire_gauge_mm2_x100)
) RETURNING
    id,
    org_id,
    equipment_type_id,
    tracking_type_id,
    category_id,
    manufacturer_id,
    usage_type_id,
    location_id,
    storage_object_id,
    name,
    is_archived,
    rental_price,
    resale_price,
    notes,
    weight_g,
    width_mm,
    height_mm,
    depth_mm,
    voltage_mv,
    current_ma,
    power_mw,
    wire_gauge_mm2_x100,
    created_at;

-- name: GetByID :one
SELECT
    e.id,
    e.org_id,
    e.equipment_type_id,
    COALESCE(et.name, '') AS equipment_type_name,
    e.tracking_type_id,
    e.category_id,
    e.manufacturer_id,
    e.usage_type_id,
    e.location_id,
    COALESCE(wl.name, '') AS location_name,
    e.storage_object_id,
    e.name,
    e.is_archived,
    CASE
        WHEN tt.name = 'bulk'       THEN COALESCE((SELECT SUM(ebi.quantity) FROM equipment_bulk_items ebi WHERE ebi.equipment_id = e.id), 0)
        WHEN tt.name = 'serialized' THEN (SELECT COUNT(*) FROM equipment_serialized_items esi WHERE esi.equipment_id = e.id)
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
    e.voltage_mv,
    e.current_ma,
    e.power_mw,
    e.wire_gauge_mm2_x100,
    e.updated_at,
    e.created_at
FROM equipment e
LEFT JOIN equipment_types et ON et.id = e.equipment_type_id
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
    voltage_mv = sqlc.arg(voltage_mv),
    current_ma = sqlc.arg(current_ma),
    power_mw = sqlc.arg(power_mw),
    wire_gauge_mm2_x100 = sqlc.arg(wire_gauge_mm2_x100),
    updated_at = unixepoch()
WHERE id = sqlc.arg(id);

-- name: UpdateStorageObject :exec
UPDATE equipment
SET storage_object_id = sqlc.arg(storage_object_id)
WHERE id = sqlc.arg(id);

-- name: Delete :exec
DELETE FROM equipment
WHERE id = ?;

-- name: Export :many
SELECT
    e.id,
    e.name,
    COALESCE(tt.name, '') AS tracking_type,
    COALESCE(ut.name, '') AS usage_type,
    COALESCE(ec.name, '') AS category,
    COALESCE(em.name, '') AS manufacturer,
    COALESCE(wl.name, '') AS location,
    COALESCE(et.name, '') AS equipment_type,
    e.rental_price,
    e.resale_price,
    e.notes,
    e.weight_g,
    e.width_mm,
    e.height_mm,
    e.depth_mm,
    e.voltage_mv,
    e.current_ma,
    e.power_mw,
    e.wire_gauge_mm2_x100
FROM equipment e
LEFT JOIN equipment_categories    ec ON ec.id = e.category_id
LEFT JOIN equipment_manufacturers em ON em.id = e.manufacturer_id
LEFT JOIN warehouse_locations     wl ON wl.id = e.location_id
LEFT JOIN equipment_types         et ON et.id = e.equipment_type_id
LEFT JOIN tracking_types          tt ON tt.id = e.tracking_type_id
LEFT JOIN usage_types             ut ON ut.id = e.usage_type_id
WHERE e.org_id = sqlc.arg(org_id)
  AND e.is_archived = 0
ORDER BY ec.name ASC, e.name ASC;

-- name: UpdateArchived :exec
UPDATE equipment
SET
    is_archived = sqlc.arg(is_archived),
    updated_at = unixepoch()
WHERE id = sqlc.arg(id);
