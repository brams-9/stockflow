package domain

import (
	"errors"
	"fmt"
)

var (
	ErrSKUExists             = errors.New("sku already exists")
	ErrProductNotFound       = errors.New("product not found")
	ErrProductInactive       = errors.New("product is inactive")
	ErrProductHasHistory     = errors.New("product has history")
	ErrOrderNotFound         = errors.New("order not found")
	ErrOrderAlreadyCancelled = errors.New("order already cancelled")
	ErrStockOverflow         = errors.New("stock would exceed the maximum")
)

type ValidationError struct {
	Message string
	Fields  map[string]string
}

func (e *ValidationError) Error() string {
	return e.Message
}

type InsufficientStockError struct {
	SKU       string
	Available int
}

func (e *InsufficientStockError) Error() string {
	return fmt.Sprintf("insufficient stock: %d available", e.Available)
}
