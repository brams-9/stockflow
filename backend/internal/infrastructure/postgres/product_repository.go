package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"stockflow/internal/domain"
	"stockflow/internal/infrastructure/postgres/sqlc"
)

const (
	uniqueViolation     = "23505"
	foreignKeyViolation = "23503"
)

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

type ProductRepository struct {
	queries *sqlc.Queries
}

func NewProductRepository(queries *sqlc.Queries) *ProductRepository {
	return &ProductRepository{queries: queries}
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
		if isSKUConflict(err) {
			return domain.Product{}, domain.ErrSKUExists
		}
		return domain.Product{}, err
	}

	return toProduct(row), nil
}

func (r *ProductRepository) GetByID(ctx context.Context, id int) (domain.Product, error) {
	row, err := r.queries.GetProductByID(ctx, int32(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Product{}, domain.ErrProductNotFound
		}
		return domain.Product{}, err
	}

	return domain.Product{
		ID:             int(row.ID),
		SKU:            row.Sku,
		Name:           row.Name,
		Price:          int(row.Price),
		Stock:          int(row.Stock),
		IsActive:       row.IsActive,
		CreatedAt:      row.CreatedAt.Time,
		LastMovementAt: optionalTime(row.LastMovementAt),
	}, nil
}

func (r *ProductRepository) FindBySKUs(ctx context.Context, skus []string) ([]domain.Product, error) {
	rows, err := r.queries.ListProductsBySKU(ctx, skus)
	if err != nil {
		return nil, err
	}

	products := make([]domain.Product, 0, len(rows))
	for _, row := range rows {
		products = append(products, toProduct(row))
	}
	return products, nil
}

func (r *ProductRepository) Update(ctx context.Context, product domain.Product) (domain.Product, error) {
	affected, err := r.queries.UpdateProduct(ctx, sqlc.UpdateProductParams{
		ID:       int32(product.ID),
		Sku:      product.SKU,
		Name:     product.Name,
		Price:    int32(product.Price),
		IsActive: product.IsActive,
	})
	if err != nil {
		if isSKUConflict(err) {
			return domain.Product{}, domain.ErrSKUExists
		}
		return domain.Product{}, err
	}
	if affected == 0 {
		return domain.Product{}, domain.ErrProductNotFound
	}

	return r.GetByID(ctx, product.ID)
}

func (r *ProductRepository) Delete(ctx context.Context, id int) error {
	hasHistory, err := r.queries.ProductHasHistory(ctx, int32(id))
	if err != nil {
		return err
	}
	if hasHistory {
		return domain.ErrProductHasHistory
	}

	affected, err := r.queries.DeleteProduct(ctx, int32(id))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == foreignKeyViolation {
			return domain.ErrProductHasHistory
		}
		return err
	}
	if affected == 0 {
		return domain.ErrProductNotFound
	}
	return nil
}

func isSKUConflict(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == uniqueViolation && pgErr.ConstraintName == "products_sku_key"
}

func toProduct(row sqlc.Product) domain.Product {
	return domain.Product{
		ID:        int(row.ID),
		SKU:       row.Sku,
		Name:      row.Name,
		Price:     int(row.Price),
		Stock:     int(row.Stock),
		IsActive:  row.IsActive,
		CreatedAt: row.CreatedAt.Time,
	}
}

func optionalTime(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}
	t := ts.Time
	return &t
}
