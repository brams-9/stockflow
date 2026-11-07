package service_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"stockflow/internal/domain"
	"stockflow/internal/infrastructure/postgres"
	"stockflow/internal/service"
)

type fixture struct {
	t        *testing.T
	pool     *pgxpool.Pool
	products *service.ProductService
	stock    *service.StockService
	orders   *service.OrderService
	prefix   string
	skus     []string
}

func newFixture(t *testing.T) *fixture {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}

	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	repos := postgres.NewRepositories(pool)
	tx := postgres.NewTxRunner(pool)
	f := &fixture{
		t:        t,
		pool:     pool,
		products: service.NewProductService(repos.Products),
		stock:    service.NewStockService(repos, tx),
		orders:   service.NewOrderService(repos, tx),
		prefix:   fmt.Sprintf("IT-%d", time.Now().UnixNano()),
	}
	t.Cleanup(f.cleanup)
	return f
}

func (f *fixture) cleanup() {
	ctx := context.Background()
	f.pool.Exec(ctx, `DELETE FROM sales_order_items WHERE sku = ANY($1)`, f.skus)
	f.pool.Exec(ctx, `DELETE FROM sales_orders WHERE customer LIKE $1`, f.prefix+"%")
	f.pool.Exec(ctx, `DELETE FROM stock_movements WHERE product_id IN (SELECT id FROM products WHERE sku = ANY($1))`, f.skus)
	f.pool.Exec(ctx, `DELETE FROM products WHERE sku = ANY($1)`, f.skus)
	f.pool.Close()
}

func (f *fixture) product(suffix string, price, stock int) domain.Product {
	f.t.Helper()
	ctx := context.Background()

	sku := f.prefix + "-" + suffix
	f.skus = append(f.skus, sku)

	product, err := f.products.Create(ctx, service.CreateProductInput{SKU: sku, Name: "Item " + suffix, Price: &price})
	if err != nil {
		f.t.Fatalf("create product: %v", err)
	}
	if stock > 0 {
		if _, err := f.stock.Adjust(ctx, product.ID, service.AdjustStockInput{Type: "in", Qty: &stock}); err != nil {
			f.t.Fatalf("seed stock: %v", err)
		}
	}
	return product
}

func (f *fixture) stockOf(sku string) int {
	f.t.Helper()
	var stock int
	if err := f.pool.QueryRow(context.Background(), `SELECT stock FROM products WHERE sku = $1`, sku).Scan(&stock); err != nil {
		f.t.Fatalf("read stock: %v", err)
	}
	return stock
}

func (f *fixture) count(query string, args ...any) int {
	f.t.Helper()
	var n int
	if err := f.pool.QueryRow(context.Background(), query, args...).Scan(&n); err != nil {
		f.t.Fatalf("count: %v", err)
	}
	return n
}

func TestOrderInsufficientStockWritesNothing(t *testing.T) {
	f := newFixture(t)
	a := f.product("A", 1000, 10)
	b := f.product("B", 2000, 1)
	customer := f.prefix + "-customer"

	_, err := f.orders.Create(context.Background(), service.CreateOrderInput{
		Customer: customer,
		Items: []service.OrderItemInput{
			{SKU: a.SKU, Qty: 2},
			{SKU: b.SKU, Qty: 5},
		},
	})

	var insufficient *domain.InsufficientStockError
	if !errors.As(err, &insufficient) {
		t.Fatalf("expected insufficient stock error, got %v", err)
	}
	if insufficient.SKU != b.SKU || insufficient.Available != 1 {
		t.Errorf("got %+v, want sku %s available 1", insufficient, b.SKU)
	}
	if got := f.stockOf(a.SKU); got != 10 {
		t.Errorf("stock of A changed to %d", got)
	}
	if got := f.stockOf(b.SKU); got != 1 {
		t.Errorf("stock of B changed to %d", got)
	}
	if n := f.count(`SELECT count(*) FROM sales_orders WHERE customer = $1`, customer); n != 0 {
		t.Errorf("%d orders were written", n)
	}
	if n := f.count(`SELECT count(*) FROM stock_movements WHERE product_id = $1 AND type = 'out'`, a.ID); n != 0 {
		t.Errorf("%d movements were written for A", n)
	}
}

func TestOrderHappyPathWritesMovementAndUpdatesStock(t *testing.T) {
	f := newFixture(t)
	a := f.product("A", 1500, 10)
	b := f.product("B", 300, 4)

	order, err := f.orders.Create(context.Background(), service.CreateOrderInput{
		Customer: f.prefix + "-customer",
		Items: []service.OrderItemInput{
			{SKU: b.SKU, Qty: 4},
			{SKU: a.SKU, Qty: 3},
		},
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}

	if order.Total != 1500*3+300*4 {
		t.Errorf("got total %d", order.Total)
	}
	if order.Status != domain.OrderCreated || len(order.Items) != 2 {
		t.Errorf("unexpected order %+v", order)
	}
	if got := f.stockOf(a.SKU); got != 7 {
		t.Errorf("stock of A is %d, want 7", got)
	}
	if got := f.stockOf(b.SKU); got != 0 {
		t.Errorf("stock of B is %d, want 0", got)
	}

	note := fmt.Sprintf("order #%d", order.ID)
	if n := f.count(`SELECT count(*) FROM stock_movements WHERE type = 'out' AND note = $1`, note); n != 2 {
		t.Errorf("got %d out movements, want 2", n)
	}
}

func TestCancelOrderRestoresStockOnce(t *testing.T) {
	f := newFixture(t)
	a := f.product("A", 1000, 10)

	order, err := f.orders.Create(context.Background(), service.CreateOrderInput{
		Customer: f.prefix + "-customer",
		Items:    []service.OrderItemInput{{SKU: a.SKU, Qty: 6}},
	})
	if err != nil {
		t.Fatalf("create order: %v", err)
	}

	cancelled, err := f.orders.Cancel(context.Background(), order.ID)
	if err != nil {
		t.Fatalf("cancel: %v", err)
	}
	if cancelled.Status != domain.OrderCancelled || cancelled.CancelledAt == nil {
		t.Errorf("unexpected cancelled order %+v", cancelled)
	}
	if got := f.stockOf(a.SKU); got != 10 {
		t.Errorf("stock is %d, want 10", got)
	}

	if _, err := f.orders.Cancel(context.Background(), order.ID); !errors.Is(err, domain.ErrOrderAlreadyCancelled) {
		t.Errorf("expected already cancelled, got %v", err)
	}
	if got := f.stockOf(a.SKU); got != 10 {
		t.Errorf("stock is %d after second cancel, want 10", got)
	}
}

func TestStockAdjustOutRejectsOverdraw(t *testing.T) {
	f := newFixture(t)
	a := f.product("A", 1000, 3)
	qty := 5

	_, err := f.stock.Adjust(context.Background(), a.ID, service.AdjustStockInput{Type: "out", Qty: &qty})

	var insufficient *domain.InsufficientStockError
	if !errors.As(err, &insufficient) || insufficient.Available != 3 {
		t.Fatalf("expected insufficient stock with 3 available, got %v", err)
	}
	if got := f.stockOf(a.SKU); got != 3 {
		t.Errorf("stock is %d, want 3", got)
	}
}

func TestConcurrentOrdersNeverOversell(t *testing.T) {
	f := newFixture(t)
	a := f.product("A", 1000, 5)

	const attempts = 20
	var wg sync.WaitGroup
	var mu sync.Mutex
	succeeded := 0
	insufficient := 0

	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := f.orders.Create(context.Background(), service.CreateOrderInput{
				Customer: fmt.Sprintf("%s-customer-%d", f.prefix, i),
				Items:    []service.OrderItemInput{{SKU: a.SKU, Qty: 1}},
			})

			mu.Lock()
			defer mu.Unlock()
			var stockErr *domain.InsufficientStockError
			switch {
			case err == nil:
				succeeded++
			case errors.As(err, &stockErr):
				insufficient++
			default:
				t.Errorf("unexpected error: %v", err)
			}
		}(i)
	}
	wg.Wait()

	if succeeded != 5 || insufficient != attempts-5 {
		t.Errorf("got %d succeeded and %d insufficient, want 5 and %d", succeeded, insufficient, attempts-5)
	}
	if got := f.stockOf(a.SKU); got != 0 {
		t.Errorf("final stock is %d, want 0", got)
	}
}
