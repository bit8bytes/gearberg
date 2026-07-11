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
    purchase_price,
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
    quantity,
    has_content,
    unit_serial_number,
    unit_manufacturer_serial,
    unit_purchase_price,
    unit_purchased_at,
    next_inspection_at,
    unit_is_active,
    unit_remark
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?, ?, ?, ?, ?, ?,
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
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

-- name: UpdateImportRowSerialNumber :exec
UPDATE equipment_imports SET unit_serial_number = ? WHERE id = ?;
