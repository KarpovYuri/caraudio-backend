package domain

import "time"

type ProductAttribute struct {
	ID        string    `db:"id"`
	ProductID string    `db:"product_id"`
	Name      string    `db:"name"`
	Value     string    `db:"value"`
	SortOrder int32     `db:"sort_order"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type ProductAttributeInput struct {
	Name      string
	Value     string
	SortOrder int32
}
