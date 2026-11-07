package repository

import (
	"context"

	"stockflow/internal/domain"
)

type OrderRepository interface {
	Create(ctx context.Context, order domain.SalesOrder) (domain.SalesOrder, error)
	GetByID(ctx context.Context, id int) (domain.SalesOrder, error)
	GetByIDForUpdate(ctx context.Context, id int) (domain.SalesOrder, error)
	MarkCancelled(ctx context.Context, id int) (domain.SalesOrder, error)
}
