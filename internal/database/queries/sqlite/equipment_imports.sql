-- name: InsertImportRow :one
INSERT INTO equipment_imports (
    id,
    import_id,
    org_id,
    row_number,
    status,
    error_message,
    action,
    existing_equipment_id,
    existing_item_id,
    name,
    type_label,
    tracking_label,
    usage_type_label,
    category_name,
    manufacturer_name,
    location_name,
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
    code,
    quantity,
    purchase_price,
    purchased_at,
    manufacturer_serial,
    last_inspected_at,
    is_active,
    remark
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?, ?, ?, ?
) RETURNING *;

-- name: DeleteImportsByOrgID :exec
DELETE FROM equipment_imports WHERE org_id = ?;

-- name: DeleteImportsByImportID :exec
DELETE FROM equipment_imports WHERE import_id = ?;

-- name: ListImportRowsByImportID :many
SELECT * FROM equipment_imports WHERE import_id = ? ORDER BY row_number ASC;

-- name: GetImportRow :one
SELECT * FROM equipment_imports WHERE id = ?;

-- name: UpdateImportRowAction :exec
UPDATE equipment_imports SET action = ? WHERE id = ?;
