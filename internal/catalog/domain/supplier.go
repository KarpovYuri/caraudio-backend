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

type SupplierInput struct {
	Name     string `json:"name"`
	Code     string `json:"code"`
	Logo     string `json:"logo"`
	ApiUrl   string `json:"api_url"`
	IsActive bool   `json:"is_active"`
}

//type (
//	SupplierListFilter struct {
//		Query      string
//		ActiveOnly bool
//		Page       int32
//		PageSize   int32
//	}
//)
//
//type SupplierListResult struct {
//	Suppliers []Supplier
//	Total     int32
//}
