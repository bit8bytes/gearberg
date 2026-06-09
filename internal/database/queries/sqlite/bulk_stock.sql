-- name: Create :exec
INSERT INTO bulk_stock (
  inventory_id, 
  quantity
) VALUES (
  sqlc.arg(inventory_id), 
  sqlc.arg(quantity)
);

-- name: SetQuantity :exec
UPDATE bulk_stock SET 
  quantity = sqlc.arg(quantity), 
  updated_at = unixepoch() WHERE inventory_id = sqlc.arg(inventory_id);
