CREATE TABLE IF NOT EXISTS supplier_product_mappings (
    id UUID PRIMARY KEY,
    product_id UUID NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    supplier_id UUID NOT NULL REFERENCES suppliers (id) ON DELETE CASCADE,
    external_id VARCHAR(255) NOT NULL,
    external_sku VARCHAR(64) NOT NULL DEFAULT '',
    external_name VARCHAR(255) NOT NULL DEFAULT '',
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (supplier_id, external_id),
    UNIQUE (product_id, supplier_id)
);

CREATE INDEX IF NOT EXISTS idx_supplier_product_mappings_supplier_id
    ON supplier_product_mappings (supplier_id);
