package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"

	"stockflow/internal/infrastructure/postgres/sqlc"
	"stockflow/internal/repository"
)

type IdempotencyRepository struct {
	queries *sqlc.Queries
}

func NewIdempotencyRepository(queries *sqlc.Queries) *IdempotencyRepository {
	return &IdempotencyRepository{queries: queries}
}

func (r *IdempotencyRepository) Claim(ctx context.Context, key, requestHash string) (bool, error) {
	affected, err := r.queries.ClaimIdempotencyKey(ctx, sqlc.ClaimIdempotencyKeyParams{
		Key:         key,
		RequestHash: requestHash,
	})
	if err != nil {
		return false, err
	}
	return affected == 1, nil
}

func (r *IdempotencyRepository) Find(ctx context.Context, key string) (repository.IdempotencyRecord, error) {
	row, err := r.queries.GetIdempotencyKey(ctx, key)
	if err != nil {
		return repository.IdempotencyRecord{}, err
	}
	if !row.OrderID.Valid {
		return repository.IdempotencyRecord{}, errors.New("idempotency key has no order")
	}
	return repository.IdempotencyRecord{RequestHash: row.RequestHash, OrderID: int(row.OrderID.Int32)}, nil
}

func (r *IdempotencyRepository) AttachOrder(ctx context.Context, key string, orderID int) error {
	return r.queries.AttachIdempotencyKeyOrder(ctx, sqlc.AttachIdempotencyKeyOrderParams{
		Key:     key,
		OrderID: pgtype.Int4{Int32: int32(orderID), Valid: true},
	})
}
