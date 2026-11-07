package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"stockflow/internal/infrastructure/postgres/sqlc"
	"stockflow/internal/repository"
)

func NewRepositories(db sqlc.DBTX) repository.Repositories {
	queries := sqlc.New(db)
	return repository.Repositories{
		Products:    NewProductRepository(queries),
		Stock:       NewStockRepository(queries),
		Movements:   NewMovementRepository(queries),
		Orders:      NewOrderRepository(queries),
		Idempotency: NewIdempotencyRepository(queries),
	}
}

type TxRunner struct {
	pool *pgxpool.Pool
}

func NewTxRunner(pool *pgxpool.Pool) *TxRunner {
	return &TxRunner{pool: pool}
}

func (r *TxRunner) WithinTx(ctx context.Context, fn func(repository.Repositories) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(NewRepositories(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
