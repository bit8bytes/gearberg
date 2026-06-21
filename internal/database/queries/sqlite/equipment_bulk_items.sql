-- name: Create :one
INSERT INTO equipment_bulk_items (
    id,
    equipment_id,
    quantity,
    purchase_price,
    purchased_at,
    remark
) VALUES (
    sqlc.arg(id),
    sqlc.arg(equipment_id),
    sqlc.arg(quantity),
    sqlc.arg(purchase_price),
    sqlc.arg(purchased_at),
    sqlc.arg(remark)
) RETURNING
    id,
    equipment_id,
    quantity,
    purchase_price,
    purchased_at,
    remark,
    created_at;

-- name: GetByEquipmentID :one
SELECT
    id,
    equipment_id,
    quantity,
    purchase_price,
    purchased_at,
    remark,
    updated_at,
    created_at
FROM equipment_bulk_items
WHERE equipment_id = ?;

-- name: SetQuantity :exec
UPDATE equipment_bulk_items
SET
    quantity = sqlc.arg(quantity),
    updated_at = unixepoch()
WHERE id = sqlc.arg(id);

-- name: Delete :exec
DELETE FROM equipment_bulk_items
WHERE id = ?;
