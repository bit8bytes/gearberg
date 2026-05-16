-- name: CreateEquipmentCategory :one
INSERT INTO equipment_categories (
    id,
    company_id,
    name
) VALUES (
    ?,
    ?,
    ?
) RETURNING
    id,
    company_id,
    name,
    created_at;

-- name: GetEquipmentCategoryByID :one
SELECT
    id,
    company_id,
    name,
    updated_at,
    created_at
FROM equipment_categories
WHERE id = ?;

-- name: GetAllEquipmentCategories :many
SELECT
    id,
    company_id,
    name,
    updated_at,
    created_at
FROM equipment_categories;

-- name: GetEquipmentCategoriesByCompanyID :many
SELECT
    id,
    company_id,
    name,
    updated_at,
    created_at
FROM equipment_categories
WHERE company_id = ?;

-- name: UpdateEquipmentCategory :one
UPDATE equipment_categories
SET
    name = ?,
    updated_at = unixepoch()
WHERE id = ?
RETURNING
    id,
    company_id,
    name,
    updated_at,
    created_at;

-- name: DeleteEquipmentCategory :exec
DELETE FROM equipment_categories
WHERE id = ?;
