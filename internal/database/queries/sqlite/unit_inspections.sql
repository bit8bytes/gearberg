-- name: CreateInspection :one
INSERT INTO unit_inspections (id, unit_id, inspected_at, passed, notes)
VALUES (sqlc.arg(id), sqlc.arg(unit_id), sqlc.arg(inspected_at), sqlc.arg(passed), sqlc.arg(notes))
RETURNING id, unit_id, inspected_at, passed, notes, created_at;

-- name: ListByUnitID :many
SELECT id, unit_id, inspected_at, passed, notes, created_at
FROM unit_inspections
WHERE unit_id = ?
ORDER BY inspected_at DESC, created_at DESC;

-- name: ListLatestInspectedAtByInventoryID :many
SELECT ui.unit_id, MAX(ui.inspected_at) AS inspected_at
FROM unit_inspections ui
INNER JOIN serialized_units su ON ui.unit_id = su.id
WHERE su.inventory_id = ?1
  AND ui.passed = 1
GROUP BY ui.unit_id;
