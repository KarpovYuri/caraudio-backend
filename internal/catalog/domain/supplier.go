package domain

import "time"

type Supplier struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Code      *string   `db:"code"`
	Logo      string    `db:"logo"`
	ApiUrl    string    `db:"api_url"`
	IsActive  bool      `db:"is_active"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
