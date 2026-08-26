-- name: Create :one
INSERT INTO equipment_serialized_items (
    id,
    org_id,
    equipment_id,
    serial_number,
    is_active,
    remark,
    purchase_price,
    purchased_at,
    next_inspection_at,
    manufacturer_serial
) VALUES (
    sqlc.arg(id),
    sqlc.arg(org_id),
    sqlc.arg(equipment_id),
    sqlc.arg(serial_number),
    sqlc.arg(is_active),
    sqlc.arg(remark),
    sqlc.arg(purchase_price),
    sqlc.arg(purchased_at),
    sqlc.arg(next_inspection_at),
    sqlc.arg(manufacturer_serial)
) RETURNING
    id,
    org_id,
    equipment_id,
    parent_item_id,
    serial_number,
    is_active,
    remark,
    purchase_price,
    purchased_at,
    next_inspection_at,
    manufacturer_serial,
    created_at;

-- name: GetByID :one
SELECT
    id,
    equipment_id,
    parent_item_id,
    serial_number,
    is_active,
    remark,
    purchase_price,
    purchased_at,
    next_inspection_at,
    manufacturer_serial,
    updated_at,
    created_at
FROM equipment_serialized_items
WHERE id = ?;

-- name: ListByEquipmentID :many
SELECT
    id,
    equipment_id,
    parent_item_id,
    serial_number,
    is_active,
    remark,
    purchase_price,
    purchased_at,
    next_inspection_at,
    manufacturer_serial,
    updated_at,
    created_at
FROM equipment_serialized_items
WHERE equipment_id = ?
ORDER BY created_at ASC;

-- name: Update :exec
UPDATE equipment_serialized_items
SET
    serial_number = sqlc.arg(serial_number),
    is_active = sqlc.arg(is_active),
    remark = sqlc.arg(remark),
    purchase_price = sqlc.arg(purchase_price),
    purchased_at = sqlc.arg(purchased_at),
    next_inspection_at = sqlc.arg(next_inspection_at),
    manufacturer_serial = sqlc.arg(manufacturer_serial),
    updated_at = unixepoch()
WHERE id = sqlc.arg(id);

-- name: UpdateNextInspectionAt :exec
UPDATE equipment_serialized_items
SET
    next_inspection_at = sqlc.arg(next_inspection_at),
    updated_at = unixepoch()
WHERE id = sqlc.arg(id);

-- name: Delete :exec
DELETE FROM equipment_serialized_items
WHERE id = ?;

-- name: ListOverdueInspections :many
SELECT
    esi.id AS unit_id,
    esi.equipment_id,
    e.name AS equipment_name,
    esi.serial_number,
    esi.next_inspection_at
FROM equipment_serialized_items esi
JOIN equipment e ON e.id = esi.equipment_id
WHERE esi.org_id = sqlc.arg(org_id)
  AND esi.next_inspection_at IS NOT NULL
  AND esi.next_inspection_at < unixepoch()
ORDER BY esi.next_inspection_at ASC
LIMIT 5;

-- name: ListSoonInspections :many
SELECT
    esi.id AS unit_id,
    esi.equipment_id,
    e.name AS equipment_name,
    esi.serial_number,
    esi.next_inspection_at
FROM equipment_serialized_items esi
JOIN equipment e ON e.id = esi.equipment_id
WHERE esi.org_id = sqlc.arg(org_id)
  AND esi.next_inspection_at IS NOT NULL
  AND esi.next_inspection_at >= unixepoch()
  AND esi.next_inspection_at <= unixepoch() + 30 * 86400
ORDER BY esi.next_inspection_at ASC
LIMIT 5;
