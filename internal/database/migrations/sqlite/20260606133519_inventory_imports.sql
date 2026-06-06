-- +goose Up
CREATE TABLE inventory_imports (
    id TEXT NOT NULL PRIMARY KEY,
    import_id TEXT NOT NULL,
    org_id TEXT NOT NULL,
    row_number INTEGER NOT NULL,
    name TEXT NOT NULL DEFAULT '',
    type_label TEXT NOT NULL DEFAULT '',
    usage_type_label TEXT NOT NULL DEFAULT '',
    category_name TEXT NOT NULL DEFAULT '',
    manufacturer_name TEXT NOT NULL DEFAULT '',
    total_stock TEXT NOT NULL DEFAULT '1',
    purchase_price TEXT NOT NULL DEFAULT '',
    rental_price TEXT NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    error_message TEXT NOT NULL DEFAULT '',
    action TEXT NOT NULL DEFAULT 'create',
    existing_item_id TEXT,
    created_at INTEGER NOT NULL DEFAULT (unixepoch())
) STRICT;

-- +goose Down
DROP TABLE inventory_imports;
