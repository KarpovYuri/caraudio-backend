package domain

import "errors"

var (
	ErrInvalidArgument     = errors.New("invalid argument")
	ErrAlreadyExists       = errors.New("already exists")
	ErrSupplierHasProducts = errors.New("supplier has products")
	ErrSupplierNotFound    = errors.New("supplier not found")
)
