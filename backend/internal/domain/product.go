package domain

import "time"

type Product struct {
	ID             int
	SKU            string
	Name           string
	Price          int
	Stock          int
	IsActive       bool
	CreatedAt      time.Time
	LastMovementAt *time.Time
}
