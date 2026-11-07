package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"stockflow/internal/domain"
	"stockflow/internal/infrastructure/postgres/sqlc"
)

type OrderRepository struct {
	queries *sqlc.Queries
}

func NewOrderRepository(queries *sqlc.Queries) *OrderRepository {
	return &OrderRepository{queries: queries}
}

func (r *OrderRepository) Create(ctx context.Context, order domain.SalesOrder) (domain.SalesOrder, error) {
	row, err := r.queries.InsertOrder(ctx, sqlc.InsertOrderParams{
		Customer: order.Customer,
		Total:    int32(order.Total),
	})
	if err != nil {
		return domain.SalesOrder{}, err
	}

	created := toOrder(row)
	created.Items = make([]domain.SalesOrderItem, 0, len(order.Items))
	for _, item := range order.Items {
		itemRow, err := r.queries.InsertOrderItem(ctx, sqlc.InsertOrderItemParams{
			OrderID:   row.ID,
			Sku:       item.SKU,
			Qty:       int32(item.Qty),
			UnitPrice: int32(item.UnitPrice),
		})
		if err != nil {
			return domain.SalesOrder{}, err
		}
		created.Items = append(created.Items, toOrderItem(itemRow))
	}
	return created, nil
}

func (r *OrderRepository) GetByID(ctx context.Context, id int) (domain.SalesOrder, error) {
	row, err := r.queries.GetOrder(ctx, int32(id))
	if err != nil {
		return domain.SalesOrder{}, orderLookupError(err)
	}
	return r.withItems(ctx, row)
}

func (r *OrderRepository) GetByIDForUpdate(ctx context.Context, id int) (domain.SalesOrder, error) {
	row, err := r.queries.GetOrderForUpdate(ctx, int32(id))
	if err != nil {
		return domain.SalesOrder{}, orderLookupError(err)
	}
	return r.withItems(ctx, row)
}

func (r *OrderRepository) MarkCancelled(ctx context.Context, id int) (domain.SalesOrder, error) {
	row, err := r.queries.CancelOrder(ctx, int32(id))
	if err != nil {
		return domain.SalesOrder{}, orderLookupError(err)
	}
	return r.withItems(ctx, row)
}

func (r *OrderRepository) withItems(ctx context.Context, row sqlc.SalesOrder) (domain.SalesOrder, error) {
	itemRows, err := r.queries.ListOrderItems(ctx, row.ID)
	if err != nil {
		return domain.SalesOrder{}, err
	}

	order := toOrder(row)
	order.Items = make([]domain.SalesOrderItem, 0, len(itemRows))
	for _, itemRow := range itemRows {
		order.Items = append(order.Items, toOrderItem(itemRow))
	}
	return order, nil
}

func orderLookupError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrOrderNotFound
	}
	return err
}

func toOrder(row sqlc.SalesOrder) domain.SalesOrder {
	return domain.SalesOrder{
		ID:          int(row.ID),
		Customer:    row.Customer,
		Status:      domain.OrderStatus(row.Status),
		Total:       int(row.Total),
		CreatedAt:   row.CreatedAt.Time,
		CancelledAt: optionalTime(row.CancelledAt),
	}
}

func toOrderItem(row sqlc.SalesOrderItem) domain.SalesOrderItem {
	return domain.SalesOrderItem{
		ID:        int(row.ID),
		OrderID:   int(row.OrderID),
		SKU:       row.Sku,
		Qty:       int(row.Qty),
		UnitPrice: int(row.UnitPrice),
	}
}
