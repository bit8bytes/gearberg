-- name: Create :one
INSERT INTO equipment_items (
    id,
    equipment_id,
    internal_id,
    is_active,
    quantity,
    remark,
    purchase_price,
    purchased_at,
    last_inspected_at,
    manufacturer_serial
) VALUES (
    sqlc.arg(id),
    sqlc.arg(equipment_id),
    sqlc.arg(internal_id),
    sqlc.arg(is_active),
    sqlc.arg(quantity),
    sqlc.arg(remark),
    sqlc.arg(purchase_price),
    sqlc.arg(purchased_at),
    sqlc.arg(last_inspected_at),
    sqlc.arg(manufacturer_serial)
) RETURNING
    id,
    equipment_id,
    parent_equipment_item_id,
    internal_id,
    is_active,
    quantity,
    remark,
    purchase_price,
    purchased_at,
    last_inspected_at,
    manufacturer_serial,
    created_at;

-- name: GetByID :one
SELECT
    id,
    equipment_id,
    parent_equipment_item_id,
    internal_id,
    is_active,
    quantity,
    remark,
    purchase_price,
    purchased_at,
    last_inspected_at,
    manufacturer_serial,
    updated_at,
    created_at
FROM equipment_items
WHERE id = ?;

-- name: ListByEquipmentID :many
SELECT
    id,
    equipment_id,
    parent_equipment_item_id,
    internal_id,
    is_active,
    quantity,
    remark,
    purchase_price,
    purchased_at,
    last_inspected_at,
    manufacturer_serial,
    updated_at,
    created_at
FROM equipment_items
WHERE equipment_id = ?
ORDER BY internal_id ASC;

-- name: Update :exec
UPDATE equipment_items
SET
    is_active = sqlc.arg(is_active),
    quantity = sqlc.arg(quantity),
    remark = sqlc.arg(remark),
    purchase_price = sqlc.arg(purchase_price),
    purchased_at = sqlc.arg(purchased_at),
    last_inspected_at = sqlc.arg(last_inspected_at),
    manufacturer_serial = sqlc.arg(manufacturer_serial),
    updated_at = unixepoch()
WHERE id = sqlc.arg(id);

-- name: SetQuantity :exec
UPDATE equipment_items
SET
    quantity = sqlc.arg(quantity),
    updated_at = unixepoch()
WHERE id = sqlc.arg(id);

-- name: Delete :exec
DELETE FROM equipment_items
WHERE id = ?;
