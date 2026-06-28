package domain

import "time"

type Category struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	Slug      string    `db:"slug"`
	ParentID  *string   `db:"parent_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
