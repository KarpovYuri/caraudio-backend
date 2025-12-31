package domain

import (
	"time"
)

type User struct {
	ID        string    `db:"id"`
	Login     string    `db:"login"`
	Password  string    `db:"password"`
	Role      string    `db:"role"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
