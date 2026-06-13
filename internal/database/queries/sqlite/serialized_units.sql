-- name: CreateUnit :one
INSERT INTO serialized_units (
    id,
    inventory_id,
    status_id,
    unit_number,
    serial_number
) VALUES (
    sqlc.arg(id),
    sqlc.arg(inventory_id),
    sqlc.arg(status_id),
    (SELECT COALESCE(MAX(unit_number), 0) + 1 FROM serialized_units WHERE inventory_id = sqlc.arg(inventory_id)),
    sqlc.arg(serial_number)
) RETURNING
    id,
    inventory_id,
    status_id,
    unit_number,
    serial_number,
    created_at;

-- name: ListByInventoryID :many
SELECT
    id,
    inventory_id,
    status_id,
    unit_number,
    serial_number,
    notes,
    purchased_at,
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
    purchased_at,
    updated_at,
    created_at
FROM serialized_units
WHERE id = ?;

-- name: Update :exec
UPDATE serialized_units
SET
    status_id = ?,
    serial_number = ?,
    notes = ?,
    purchased_at = ?,
    updated_at = unixepoch()
WHERE id = ?;

-- name: Delete :exec
DELETE FROM serialized_units
WHERE id = ?;
