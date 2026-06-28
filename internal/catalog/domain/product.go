package domain

import "time"

type Product struct {
	ID          string    `db:"id"`
	CategoryID  *string   `db:"category_id"`
	BrandID     *string   `db:"brand_id"`
	SupplierID  *int64    `db:"supplier_id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	PriceCents  int64     `db:"price_cents"`
	SKU         *string   `db:"sku"`
	Stock       int32     `db:"stock"`
	IsActive    bool      `db:"is_active"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type ProductListFilter struct {
	CategoryID string
	BrandID    string
	SupplierID int64
	ActiveOnly bool
	Page       int32
	PageSize   int32
}

type ProductListResult struct {
	Products []Product
	Total    int32
}
