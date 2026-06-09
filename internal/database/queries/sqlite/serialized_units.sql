-- name: CreateUnit :one
INSERT INTO serialized_units (
    id,
    inventory_id,
    status_id,
    unit_number,
    serial_number,
    next_inspection_at
) VALUES (
    sqlc.arg(id),
    sqlc.arg(inventory_id),
    sqlc.arg(status_id),
    (SELECT COALESCE(MAX(unit_number), 0) + 1 FROM serialized_units WHERE inventory_id = sqlc.arg(inventory_id)),
    sqlc.arg(serial_number),
    sqlc.arg(next_inspection_at)
) RETURNING
    id,
    inventory_id,
    status_id,
    unit_number,
    serial_number,
    next_inspection_at,
    created_at;

-- name: ListByInventoryID :many
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
FROM serialized_units
WHERE inventory_id = ?
ORDER BY unit_number ASC;

-- name: GetByID :one
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
FROM serialized_units
WHERE id = ?;

-- name: Update :exec
UPDATE serialized_units
SET
    status_id = ?,
    serial_number = ?,
    next_inspection_at = ?,
    notes = ?,
    updated_at = unixepoch()
WHERE id = ?;

-- name: Delete :exec
DELETE FROM serialized_units
WHERE id = ?;
