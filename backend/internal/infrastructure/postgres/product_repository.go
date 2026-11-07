package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"stockflow/internal/domain"
	"stockflow/internal/infrastructure/postgres/sqlc"
)

const uniqueViolation = "23505"

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

type ProductRepository struct {
	queries *sqlc.Queries
}

func NewProductRepository(pool *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{queries: sqlc.New(pool)}
}

func (r *ProductRepository) List(ctx context.Context, search string, limit, offset int) ([]domain.Product, int, error) {
	pattern := "%" + likeEscaper.Replace(search) + "%"

	rows, err := r.queries.ListProducts(ctx, sqlc.ListProductsParams{
		Pattern:   pattern,
		RowLimit:  int32(limit),
		RowOffset: int32(offset),
	})
	if err != nil {
		return nil, 0, err
	}

	total, err := r.queries.CountProducts(ctx, pattern)
	if err != nil {
		return nil, 0, err
	}

	products := make([]domain.Product, 0, len(rows))
	for _, row := range rows {
		products = append(products, domain.Product{
			ID:             int(row.ID),
			SKU:            row.Sku,
			Name:           row.Name,
			Price:          int(row.Price),
			Stock:          int(row.Stock),
			IsActive:       row.IsActive,
			CreatedAt:      row.CreatedAt.Time,
			LastMovementAt: optionalTime(row.LastMovementAt),
		})
	}
	return products, int(total), nil
}

func (r *ProductRepository) Create(ctx context.Context, product domain.Product) (domain.Product, error) {
	row, err := r.queries.CreateProduct(ctx, sqlc.CreateProductParams{
		Sku:   product.SKU,
		Name:  product.Name,
		Price: int32(product.Price),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation && pgErr.ConstraintName == "products_sku_key" {
			return domain.Product{}, domain.ErrSKUExists
		}
		return domain.Product{}, err
	}

	return domain.Product{
		ID:        int(row.ID),
		SKU:       row.Sku,
		Name:      row.Name,
		Price:     int(row.Price),
		Stock:     int(row.Stock),
		IsActive:  row.IsActive,
		CreatedAt: row.CreatedAt.Time,
	}, nil
}

func optionalTime(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}
	t := ts.Time
	return &t
}
