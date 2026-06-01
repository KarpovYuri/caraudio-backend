package domain

import "errors"

var (
	ErrAlreadyExists    = errors.New("already exists")
	ErrSupplierNotFound = errors.New("supplier not found")
)
