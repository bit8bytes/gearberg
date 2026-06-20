-- name: ListByEquipmentID :many
SELECT
    eci.id,
    eci.equipment_id,
    eci.member_equipment_id,
    eci.quantity,
    e.name AS member_name,
    e.tracking_type_id AS member_tracking_type_id
FROM equipment_combination_items eci
JOIN equipment e ON e.id = eci.member_equipment_id
WHERE eci.equipment_id = ?
ORDER BY e.name ASC;

-- name: Create :one
INSERT INTO equipment_combination_items (
    id,
    equipment_id,
    member_equipment_id,
    quantity
) VALUES (
    sqlc.arg(id),
    sqlc.arg(equipment_id),
    sqlc.arg(member_equipment_id),
    sqlc.arg(quantity)
) RETURNING id, equipment_id, member_equipment_id, quantity;

-- name: Delete :exec
DELETE FROM equipment_combination_items
WHERE id = ?;

-- name: ListContainersByMemberID :many
SELECT e.id, e.name
FROM equipment e
JOIN equipment_combination_items eci ON eci.equipment_id = e.id
WHERE eci.member_equipment_id = ?
ORDER BY e.name ASC;

-- name: TotalDemandByMemberID :one
SELECT COALESCE(SUM(
    eci.quantity * (
        CASE
            WHEN tt.name = 'bulk'
                THEN COALESCE((SELECT SUM(ei.quantity) FROM equipment_items ei WHERE ei.equipment_id = eci.equipment_id AND ei.parent_equipment_item_id IS NULL), 0)
            WHEN tt.name = 'serialized'
                THEN (SELECT COUNT(*) FROM equipment_items ei WHERE ei.equipment_id = eci.equipment_id)
            ELSE 0
        END
    )
), 0) AS total_demand
FROM equipment_combination_items eci
JOIN equipment e ON e.id = eci.equipment_id
JOIN tracking_types tt ON tt.id = e.tracking_type_id
WHERE eci.member_equipment_id = ?;
