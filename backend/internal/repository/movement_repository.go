package repository

import (
	"context"

	"stockflow/internal/domain"
)

type MovementRepository interface {
	Create(ctx context.Context, movement domain.StockMovement) (domain.StockMovement, error)
	ListByProduct(ctx context.Context, productID int) ([]domain.StockMovement, error)
}
