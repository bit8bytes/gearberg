-- name: InsertImportRow :one
INSERT INTO inventory_imports (
    id,
    import_id,
    org_id,
    row_number,
    name,
    type_label,
    usage_type_label,
    category_name,
    manufacturer_name,
    total_stock,
    purchase_price,
    rental_price,
    notes,
    status,
    error_message,
    action,
    existing_item_id
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
) RETURNING *;

-- name: DeleteImportsByOrgID :exec
DELETE FROM inventory_imports WHERE org_id = ?;

-- name: DeleteImportsByImportID :exec
DELETE FROM inventory_imports WHERE import_id = ?;

-- name: ListImportRowsByImportID :many
SELECT * FROM inventory_imports WHERE import_id = ? ORDER BY row_number ASC;

-- name: GetImportRow :one
SELECT * FROM inventory_imports WHERE id = ?;

-- name: UpdateImportRowAction :exec
UPDATE inventory_imports SET action = ? WHERE id = ?;
