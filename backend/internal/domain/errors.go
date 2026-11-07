package domain

import "errors"

var ErrSKUExists = errors.New("sku already exists")

type ValidationError struct {
	Message string
	Fields  map[string]string
}

func (e *ValidationError) Error() string {
	return e.Message
}
