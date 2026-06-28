package domain

import "time"

type SupplierCategoryMapping struct {
	ID           string    `db:"id"`
	CategoryID   string    `db:"category_id"`
	SupplierID   int64     `db:"supplier_id"`
	ExternalID   string    `db:"external_id"`
	ExternalName string    `db:"external_name"`
	Notes        string    `db:"notes"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type SupplierProductMapping struct {
	ID           string    `db:"id"`
	ProductID    string    `db:"product_id"`
	SupplierID   int64     `db:"supplier_id"`
	ExternalID   string    `db:"external_id"`
	ExternalSKU  string    `db:"external_sku"`
	ExternalName string    `db:"external_name"`
	Notes        string    `db:"notes"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type SupplierCategoryMappingInput struct {
	CategoryID   string
	SupplierID   int64
	ExternalID   string
	ExternalName string
	Notes        string
}

type SupplierProductMappingInput struct {
	ProductID    string
	SupplierID   int64
	ExternalID   string
	ExternalSKU  string
	ExternalName string
	Notes        string
}

type SupplierCategoryMappingFilter struct {
	SupplierID int64
	CategoryID string
}

type SupplierProductMappingFilter struct {
	SupplierID int64
	ProductID  string
}
