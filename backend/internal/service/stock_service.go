package service

import (
	"context"
	"errors"
	"math"
	"strings"

	"stockflow/internal/domain"
	"stockflow/internal/repository"
)

type AdjustStockInput struct {
	Type string
	Qty  *int
	Note *string
}

type StockService struct {
	repos repository.Repositories
	tx    repository.TxRunner
}

func NewStockService(repos repository.Repositories, tx repository.TxRunner) *StockService {
	return &StockService{repos: repos, tx: tx}
}

func (s *StockService) Adjust(ctx context.Context, productID int, input AdjustStockInput) (domain.StockMovement, error) {
	movementType := domain.MovementType(input.Type)

	fields := map[string]string{}
	if movementType != domain.MovementIn && movementType != domain.MovementOut {
		fields["type"] = "must be 'in' or 'out'"
	}
	switch {
	case input.Qty == nil:
		fields["qty"] = "required"
	case *input.Qty <= 0:
		fields["qty"] = "must be a positive integer"
	case *input.Qty > math.MaxInt32:
		fields["qty"] = "must not exceed 2147483647"
	}
	if len(fields) > 0 {
		return domain.StockMovement{}, &domain.ValidationError{Message: "validation failed", Fields: fields}
	}

	var note *string
	if input.Note != nil {
		if trimmed := strings.TrimSpace(*input.Note); trimmed != "" {
			note = &trimmed
		}
	}
	qty := *input.Qty

	var movement domain.StockMovement
	err := s.tx.WithinTx(ctx, func(r repository.Repositories) error {
		var err error
		if movementType == domain.MovementOut {
			_, err = r.Stock.Decrease(ctx, productID, qty)
		} else {
			_, err = r.Stock.Increase(ctx, productID, qty)
		}
		if err != nil {
			return err
		}

		movement, err = r.Movements.Create(ctx, domain.StockMovement{
			ProductID: productID,
			Type:      movementType,
			Qty:       qty,
			Note:      note,
		})
		return err
	})
	if errors.Is(err, domain.ErrStockOverflow) {
		return domain.StockMovement{}, &domain.ValidationError{
			Message: "validation failed",
			Fields:  map[string]string{"qty": "would exceed the maximum stock"},
		}
	}
	if err != nil {
		return domain.StockMovement{}, err
	}
	return movement, nil
}

func (s *StockService) Movements(ctx context.Context, productID int) ([]domain.StockMovement, error) {
	if _, err := s.repos.Products.GetByID(ctx, productID); err != nil {
		return nil, err
	}
	return s.repos.Movements.ListByProduct(ctx, productID)
}
