package domain

import "errors"

var (
	ErrInvalidArgument          = errors.New("invalid argument")
	ErrSupplierHasProducts      = errors.New("supplier has products")
	ErrNotFound                 = errors.New("not found")
	ErrAlreadyExists            = errors.New("already exists")
	ErrCategoryNotFound         = errors.New("category not found")
	ErrProductNotFound          = errors.New("product not found")
	ErrSupplierNotFound         = errors.New("supplier not found")
	ErrProductImageNotFound     = errors.New("product image not found")
	ErrBrandNotFound            = errors.New("brand not found")
	ErrBrandHasProducts         = errors.New("brand has products")
	ErrProductAttributeNotFound = errors.New("product attribute not found")
	ErrSupplierMappingNotFound  = errors.New("supplier mapping not found")
	ErrCategoryHasProducts      = errors.New("category has products")
)
