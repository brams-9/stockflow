package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"stockflow/internal/domain"
	"stockflow/internal/infrastructure/postgres/sqlc"
	"stockflow/internal/repository"
)

type StockRepository struct {
	queries *sqlc.Queries
}

func NewStockRepository(queries *sqlc.Queries) *StockRepository {
	return &StockRepository{queries: queries}
}

func (r *StockRepository) Increase(ctx context.Context, productID, qty int) (int, error) {
	stock, err := r.queries.IncreaseStock(ctx, sqlc.IncreaseStockParams{ID: int32(productID), Qty: int32(qty)})
	if err == nil {
		return int(stock), nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}

	if _, err := r.queries.GetProductStock(ctx, int32(productID)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, domain.ErrProductNotFound
		}
		return 0, err
	}
	return 0, domain.ErrStockOverflow
}

func (r *StockRepository) Decrease(ctx context.Context, productID, qty int) (int, error) {
	stock, err := r.queries.DecreaseStock(ctx, sqlc.DecreaseStockParams{ID: int32(productID), Qty: int32(qty)})
	if err == nil {
		return int(stock), nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}

	available, err := r.queries.GetProductStock(ctx, int32(productID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, domain.ErrProductNotFound
		}
		return 0, err
	}
	return 0, &domain.InsufficientStockError{Available: int(available)}
}

func (r *StockRepository) DecreaseBySKU(ctx context.Context, sku string, qty int) (repository.Reservation, error) {
	row, err := r.queries.DecreaseStockBySKU(ctx, sqlc.DecreaseStockBySKUParams{Sku: sku, Qty: int32(qty)})
	if err == nil {
		return repository.Reservation{ProductID: int(row.ID), UnitPrice: int(row.Price)}, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return repository.Reservation{}, err
	}

	state, err := r.queries.GetProductStateBySKU(ctx, sku)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return repository.Reservation{}, domain.ErrProductNotFound
		}
		return repository.Reservation{}, err
	}
	if !state.IsActive {
		return repository.Reservation{}, domain.ErrProductInactive
	}
	return repository.Reservation{}, &domain.InsufficientStockError{SKU: sku, Available: int(state.Stock)}
}

func (r *StockRepository) IncreaseBySKU(ctx context.Context, sku string, qty int) (int, error) {
	id, err := r.queries.IncreaseStockBySKU(ctx, sqlc.IncreaseStockBySKUParams{Sku: sku, Qty: int32(qty)})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, domain.ErrProductNotFound
		}
		return 0, err
	}
	return int(id), nil
}
