package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"stockflow/internal/domain"
	"stockflow/internal/repository"
)

type memStore struct {
	stock       map[string]int
	prices      map[string]int
	keys        map[string]repository.IdempotencyRecord
	orders      map[int]domain.SalesOrder
	movements   int
	nextOrderID int
}

func newMemStore() *memStore {
	return &memStore{
		stock:       map[string]int{"A": 10, "B": 5},
		prices:      map[string]int{"A": 1000, "B": 2000},
		keys:        map[string]repository.IdempotencyRecord{},
		orders:      map[int]domain.SalesOrder{},
		nextOrderID: 1,
	}
}

func (m *memStore) clone() *memStore {
	c := &memStore{
		stock:       map[string]int{},
		prices:      m.prices,
		keys:        map[string]repository.IdempotencyRecord{},
		orders:      map[int]domain.SalesOrder{},
		movements:   m.movements,
		nextOrderID: m.nextOrderID,
	}
	for k, v := range m.stock {
		c.stock[k] = v
	}
	for k, v := range m.keys {
		c.keys[k] = v
	}
	for k, v := range m.orders {
		c.orders[k] = v
	}
	return c
}

type memProducts struct{ store *memStore }

func (r memProducts) List(ctx context.Context, search string, limit, offset int) ([]domain.Product, int, error) {
	return nil, 0, nil
}

func (r memProducts) Create(ctx context.Context, product domain.Product) (domain.Product, error) {
	return product, nil
}

func (r memProducts) GetByID(ctx context.Context, id int) (domain.Product, error) {
	return domain.Product{}, domain.ErrProductNotFound
}

func (r memProducts) FindBySKUs(ctx context.Context, skus []string) ([]domain.Product, error) {
	products := []domain.Product{}
	for _, sku := range skus {
		if _, ok := r.store.stock[sku]; ok {
			products = append(products, domain.Product{SKU: sku, IsActive: true})
		}
	}
	return products, nil
}

func (r memProducts) Update(ctx context.Context, product domain.Product) (domain.Product, error) {
	return product, nil
}

func (r memProducts) Delete(ctx context.Context, id int) error {
	return nil
}

type memStock struct{ store *memStore }

func (r memStock) Increase(ctx context.Context, productID, qty int) (int, error) {
	return 0, nil
}

func (r memStock) Decrease(ctx context.Context, productID, qty int) (int, error) {
	return 0, nil
}

func (r memStock) DecreaseBySKU(ctx context.Context, sku string, qty int) (repository.Reservation, error) {
	available := r.store.stock[sku]
	if available < qty {
		return repository.Reservation{}, &domain.InsufficientStockError{SKU: sku, Available: available}
	}
	r.store.stock[sku] = available - qty
	return repository.Reservation{ProductID: 1, UnitPrice: r.store.prices[sku]}, nil
}

func (r memStock) IncreaseBySKU(ctx context.Context, sku string, qty int) (int, error) {
	r.store.stock[sku] += qty
	return 1, nil
}

type memMovements struct{ store *memStore }

func (r memMovements) Create(ctx context.Context, movement domain.StockMovement) (domain.StockMovement, error) {
	r.store.movements++
	return movement, nil
}

func (r memMovements) ListByProduct(ctx context.Context, productID int) ([]domain.StockMovement, error) {
	return nil, nil
}

type memOrders struct{ store *memStore }

func (r memOrders) Create(ctx context.Context, order domain.SalesOrder) (domain.SalesOrder, error) {
	order.ID = r.store.nextOrderID
	r.store.nextOrderID++
	order.Status = domain.OrderCreated
	for i := range order.Items {
		order.Items[i].ID = i + 1
		order.Items[i].OrderID = order.ID
	}
	r.store.orders[order.ID] = order
	return order, nil
}

func (r memOrders) GetByID(ctx context.Context, id int) (domain.SalesOrder, error) {
	order, ok := r.store.orders[id]
	if !ok {
		return domain.SalesOrder{}, domain.ErrOrderNotFound
	}
	return order, nil
}

func (r memOrders) GetByIDForUpdate(ctx context.Context, id int) (domain.SalesOrder, error) {
	return r.GetByID(ctx, id)
}

func (r memOrders) MarkCancelled(ctx context.Context, id int) (domain.SalesOrder, error) {
	return r.GetByID(ctx, id)
}

type memKeys struct{ store *memStore }

func (r memKeys) Claim(ctx context.Context, key, requestHash string) (bool, error) {
	if _, exists := r.store.keys[key]; exists {
		return false, nil
	}
	r.store.keys[key] = repository.IdempotencyRecord{RequestHash: requestHash}
	return true, nil
}

func (r memKeys) Find(ctx context.Context, key string) (repository.IdempotencyRecord, error) {
	return r.store.keys[key], nil
}

func (r memKeys) AttachOrder(ctx context.Context, key string, orderID int) error {
	record := r.store.keys[key]
	record.OrderID = orderID
	r.store.keys[key] = record
	return nil
}

type memTx struct{ store **memStore }

func (t memTx) WithinTx(ctx context.Context, fn func(repository.Repositories) error) error {
	snapshot := (*t.store).clone()
	working := *t.store
	err := fn(repository.Repositories{
		Products:    memProducts{working},
		Stock:       memStock{working},
		Movements:   memMovements{working},
		Orders:      memOrders{working},
		Idempotency: memKeys{working},
	})
	if err != nil {
		*t.store = snapshot
	}
	return err
}

func newIdempotencyService() (*OrderService, **memStore) {
	store := newMemStore()
	ref := &store
	return NewOrderService(repository.Repositories{Orders: memOrders{store}}, memTx{ref}), ref
}

func TestCreateIdempotentReplaysSameOrder(t *testing.T) {
	svc, store := newIdempotencyService()
	ctx := context.Background()

	first, replayed, err := svc.CreateIdempotent(ctx, "key-1", CreateOrderInput{
		Customer: "Budi",
		Items:    []OrderItemInput{{SKU: "A", Qty: 2}, {SKU: "B", Qty: 1}},
	})
	if err != nil || replayed {
		t.Fatalf("first call: replayed=%v err=%v", replayed, err)
	}

	second, replayed, err := svc.CreateIdempotent(ctx, " key-1 ", CreateOrderInput{
		Customer: " Budi ",
		Items:    []OrderItemInput{{SKU: "B", Qty: 1}, {SKU: "A", Qty: 2}},
	})
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if !replayed {
		t.Error("expected the second call to be a replay")
	}
	if second.ID != first.ID || second.Total != first.Total {
		t.Errorf("got order %+v, want %+v", second, first)
	}
	if len((*store).orders) != 1 {
		t.Errorf("got %d orders, want 1", len((*store).orders))
	}
	if (*store).stock["A"] != 8 || (*store).stock["B"] != 4 {
		t.Errorf("stock decremented more than once: %v", (*store).stock)
	}
}

func TestCreateIdempotentRejectsDifferentPayload(t *testing.T) {
	svc, store := newIdempotencyService()
	ctx := context.Background()

	if _, _, err := svc.CreateIdempotent(ctx, "key-1", CreateOrderInput{
		Customer: "Budi",
		Items:    []OrderItemInput{{SKU: "A", Qty: 2}},
	}); err != nil {
		t.Fatalf("first call: %v", err)
	}

	_, _, err := svc.CreateIdempotent(ctx, "key-1", CreateOrderInput{
		Customer: "Budi",
		Items:    []OrderItemInput{{SKU: "A", Qty: 3}},
	})
	if !errors.Is(err, domain.ErrIdempotencyKeyReused) {
		t.Fatalf("expected ErrIdempotencyKeyReused, got %v", err)
	}
	if len((*store).orders) != 1 || (*store).stock["A"] != 8 {
		t.Errorf("state changed: orders=%d stock=%v", len((*store).orders), (*store).stock)
	}
}

func TestCreateIdempotentFailedAttemptKeepsKeyFree(t *testing.T) {
	svc, store := newIdempotencyService()
	ctx := context.Background()

	_, _, err := svc.CreateIdempotent(ctx, "key-1", CreateOrderInput{
		Customer: "Budi",
		Items:    []OrderItemInput{{SKU: "A", Qty: 1}, {SKU: "B", Qty: 50}},
	})
	var insufficient *domain.InsufficientStockError
	if !errors.As(err, &insufficient) {
		t.Fatalf("expected insufficient stock, got %v", err)
	}
	if len((*store).keys) != 0 {
		t.Errorf("failed attempt stored %d keys", len((*store).keys))
	}
	if (*store).stock["A"] != 10 {
		t.Errorf("stock of A changed to %d", (*store).stock["A"])
	}

	order, replayed, err := svc.CreateIdempotent(ctx, "key-1", CreateOrderInput{
		Customer: "Budi",
		Items:    []OrderItemInput{{SKU: "A", Qty: 1}, {SKU: "B", Qty: 1}},
	})
	if err != nil || replayed {
		t.Fatalf("retry: replayed=%v err=%v", replayed, err)
	}
	if order.ID == 0 || (*store).stock["A"] != 9 {
		t.Errorf("retry did not create the order: %+v stock=%v", order, (*store).stock)
	}
}

func TestCreateIdempotentKeyValidation(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"blank", "   "},
		{"empty", ""},
		{"too long", strings.Repeat("k", 256)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, store := newIdempotencyService()

			_, _, err := svc.CreateIdempotent(context.Background(), tt.key, CreateOrderInput{
				Customer: "Budi",
				Items:    []OrderItemInput{{SKU: "A", Qty: 1}},
			})

			var validationErr *domain.ValidationError
			if !errors.As(err, &validationErr) || validationErr.Fields["Idempotency-Key"] == "" {
				t.Fatalf("expected key validation error, got %v", err)
			}
			if len((*store).orders) != 0 {
				t.Error("an order was created")
			}
		})
	}
}

func TestCreateIdempotentAcceptsMaxLengthKey(t *testing.T) {
	svc, _ := newIdempotencyService()

	_, _, err := svc.CreateIdempotent(context.Background(), strings.Repeat("k", 255), CreateOrderInput{
		Customer: "Budi",
		Items:    []OrderItemInput{{SKU: "A", Qty: 1}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
