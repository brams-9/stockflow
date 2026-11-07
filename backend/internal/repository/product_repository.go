package repository

import (
	"context"

	"stockflow/internal/domain"
)

type ProductRepository interface {
	List(ctx context.Context, search string, limit, offset int) ([]domain.Product, int, error)
	Create(ctx context.Context, product domain.Product) (domain.Product, error)
	GetByID(ctx context.Context, id int) (domain.Product, error)
	FindBySKUs(ctx context.Context, skus []string) ([]domain.Product, error)
	Update(ctx context.Context, product domain.Product) (domain.Product, error)
	Delete(ctx context.Context, id int) error
}
