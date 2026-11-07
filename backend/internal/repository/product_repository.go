package repository

import (
	"context"

	"stockflow/internal/domain"
)

type ProductRepository interface {
	List(ctx context.Context, search string, limit, offset int) ([]domain.Product, int, error)
	Create(ctx context.Context, product domain.Product) (domain.Product, error)
}
