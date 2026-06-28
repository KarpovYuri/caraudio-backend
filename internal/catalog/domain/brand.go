package domain

import "time"

type Brand struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	Slug        string    `db:"slug"`
	Description string    `db:"description"`
	IsActive    bool      `db:"is_active"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

type BrandInput struct {
	Name        string
	Slug        string
	Description string
	IsActive    bool
}
