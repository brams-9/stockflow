package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"stockflow/internal/domain"
	"stockflow/internal/infrastructure/postgres/sqlc"
)

type MovementRepository struct {
	queries *sqlc.Queries
}

func NewMovementRepository(queries *sqlc.Queries) *MovementRepository {
	return &MovementRepository{queries: queries}
}

func (r *MovementRepository) Create(ctx context.Context, movement domain.StockMovement) (domain.StockMovement, error) {
	note := pgtype.Text{}
	if movement.Note != nil {
		note = pgtype.Text{String: *movement.Note, Valid: true}
	}

	row, err := r.queries.InsertStockMovement(ctx, sqlc.InsertStockMovementParams{
		ProductID: int32(movement.ProductID),
		Type:      string(movement.Type),
		Qty:       int32(movement.Qty),
		Note:      note,
	})
	if err != nil {
		return domain.StockMovement{}, err
	}
	return toMovement(row), nil
}

func (r *MovementRepository) ListByProduct(ctx context.Context, productID int) ([]domain.StockMovement, error) {
	rows, err := r.queries.ListStockMovements(ctx, int32(productID))
	if err != nil {
		return nil, err
	}

	movements := make([]domain.StockMovement, 0, len(rows))
	for _, row := range rows {
		movements = append(movements, toMovement(row))
	}
	return movements, nil
}

func toMovement(row sqlc.StockMovement) domain.StockMovement {
	movement := domain.StockMovement{
		ID:        int(row.ID),
		ProductID: int(row.ProductID),
		Type:      domain.MovementType(row.Type),
		Qty:       int(row.Qty),
		CreatedAt: row.CreatedAt.Time,
	}
	if row.Note.Valid {
		note := row.Note.String
		movement.Note = &note
	}
	return movement
}
