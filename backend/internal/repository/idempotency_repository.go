package repository

import "context"

type IdempotencyRecord struct {
	RequestHash string
	OrderID     int
}

type IdempotencyRepository interface {
	Claim(ctx context.Context, key, requestHash string) (bool, error)
	Find(ctx context.Context, key string) (IdempotencyRecord, error)
	AttachOrder(ctx context.Context, key string, orderID int) error
}
