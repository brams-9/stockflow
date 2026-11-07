package repository

import "context"

type Repositories struct {
	Products  ProductRepository
	Stock     StockRepository
	Movements MovementRepository
	Orders    OrderRepository
}

type TxRunner interface {
	WithinTx(ctx context.Context, fn func(Repositories) error) error
}
