CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY,
    category_id UUID REFERENCES categories (id) ON DELETE SET NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    price_cents BIGINT NOT NULL CHECK (price_cents >= 0),
    sku VARCHAR(64) UNIQUE,
    stock INT NOT NULL DEFAULT 0 CHECK (stock >= 0),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_products_category_id ON products (category_id);
CREATE INDEX IF NOT EXISTS idx_products_is_active ON products (is_active);
