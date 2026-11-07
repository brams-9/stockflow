package domain

import "time"

type MovementType string

const (
	MovementIn  MovementType = "in"
	MovementOut MovementType = "out"
)

type StockMovement struct {
	ID        int
	ProductID int
	Type      MovementType
	Qty       int
	Note      *string
	CreatedAt time.Time
}
