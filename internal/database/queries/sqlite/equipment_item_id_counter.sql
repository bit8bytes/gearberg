-- name: NextInternalID :one
INSERT INTO equipment_item_id_counter (org_id, seq)
VALUES (?, 1)
ON CONFLICT (org_id) DO UPDATE SET seq = seq + 1
RETURNING seq;
