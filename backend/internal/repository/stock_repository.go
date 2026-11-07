package repository

import "context"

type Reservation struct {
	ProductID int
	UnitPrice int
}

type StockRepository interface {
	Increase(ctx context.Context, productID, qty int) (int, error)
	Decrease(ctx context.Context, productID, qty int) (int, error)
	DecreaseBySKU(ctx context.Context, sku string, qty int) (Reservation, error)
	IncreaseBySKU(ctx context.Context, sku string, qty int) (int, error)
}
