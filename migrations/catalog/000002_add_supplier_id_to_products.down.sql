DROP INDEX IF EXISTS idx_products_supplier_id;
ALTER TABLE products DROP COLUMN IF EXISTS supplier_id;
