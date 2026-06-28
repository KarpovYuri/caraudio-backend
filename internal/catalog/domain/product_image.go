package domain

import "time"

type ProductImage struct {
	ID        string    `db:"id"`
	ProductID string    `db:"product_id"`
	URL       string    `db:"url"`
	AltText   string    `db:"alt_text"`
	SortOrder int32     `db:"sort_order"`
	IsPrimary bool      `db:"is_primary"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type ProductImageInput struct {
	URL       string
	AltText   string
	SortOrder int32
	IsPrimary bool
}
