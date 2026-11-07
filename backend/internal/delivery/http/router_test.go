package http

import (
	"context"
	"encoding/json"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"stockflow/internal/delivery/http/handler"
	"stockflow/internal/domain"
	"stockflow/internal/repository"
	"stockflow/internal/service"
)

type stubProductRepository struct{}

func (stubProductRepository) List(ctx context.Context, search string, limit, offset int) ([]domain.Product, int, error) {
	return []domain.Product{}, 0, nil
}

func (stubProductRepository) GetByID(ctx context.Context, id int) (domain.Product, error) {
	return domain.Product{}, domain.ErrProductNotFound
}

func (stubProductRepository) FindBySKUs(ctx context.Context, skus []string) ([]domain.Product, error) {
	return []domain.Product{}, nil
}

func (stubProductRepository) Update(ctx context.Context, product domain.Product) (domain.Product, error) {
	return domain.Product{}, domain.ErrProductNotFound
}

func (stubProductRepository) Delete(ctx context.Context, id int) error {
	return domain.ErrProductNotFound
}

func (stubProductRepository) Create(ctx context.Context, product domain.Product) (domain.Product, error) {
	product.ID = 7
	product.IsActive = true
	product.CreatedAt = time.Date(2026, 11, 7, 10, 0, 0, 0, time.UTC)
	return product, nil
}

func serve(method, path, body string) *httptest.ResponseRecorder {
	return serveWithHeaders(method, path, body, nil)
}

func serveWithHeaders(method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	repos := repository.Repositories{Products: stubProductRepository{}}
	router := NewRouter(
		handler.NewProductHandler(service.NewProductService(repos.Products)),
		handler.NewStockHandler(service.NewStockService(repos, nil)),
		handler.NewOrderHandler(service.NewOrderService(repos, nil)),
	)

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for name, value := range headers {
		req.Header.Set(name, value)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestCreateProductReturns201(t *testing.T) {
	rec := serve(nethttp.MethodPost, "/api/products", `{"sku":"RC-1","name":"Rice","price":1500}`)

	if rec.Code != nethttp.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if body["sku"] != "RC-1" || body["stock"] != float64(0) || body["is_active"] != true {
		t.Errorf("unexpected body: %v", body)
	}
	if value, ok := body["last_movement_at"]; !ok || value != nil {
		t.Errorf("expected last_movement_at null, got %v", value)
	}
}

func TestCreateProductValidationShape(t *testing.T) {
	rec := serve(nethttp.MethodPost, "/api/products", `{"name":"","price":-1}`)

	if rec.Code != nethttp.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}

	var body struct {
		Error struct {
			Code    string            `json:"code"`
			Message string            `json:"message"`
			Fields  map[string]string `json:"fields"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if body.Error.Code != "VALIDATION_ERROR" || len(body.Error.Fields) != 3 {
		t.Errorf("unexpected error body: %+v", body.Error)
	}
}

func TestListProductsEmptyDataIsArray(t *testing.T) {
	rec := serve(nethttp.MethodGet, "/api/products", "")

	if rec.Code != nethttp.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"data":[]`) {
		t.Errorf("expected empty data array, got %s", rec.Body.String())
	}
}

func errorCode(t *testing.T, rec *httptest.ResponseRecorder) (string, map[string]string) {
	t.Helper()
	var body struct {
		Error struct {
			Code   string            `json:"code"`
			Fields map[string]string `json:"fields"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	return body.Error.Code, body.Error.Fields
}

func TestInvalidPathIDIsValidationError(t *testing.T) {
	for _, path := range []string{"/api/products/abc/movements", "/api/products/0/movements", "/api/orders/-1"} {
		rec := serve(nethttp.MethodGet, path, "")
		if rec.Code != nethttp.StatusBadRequest {
			t.Fatalf("%s: expected 400, got %d", path, rec.Code)
		}
		code, fields := errorCode(t, rec)
		if code != "VALIDATION_ERROR" || fields["id"] == "" {
			t.Errorf("%s: unexpected error %s %v", path, code, fields)
		}
	}
}

func TestMissingProductReturns404(t *testing.T) {
	rec := serve(nethttp.MethodDelete, "/api/products/5", "")
	if rec.Code != nethttp.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	if code, _ := errorCode(t, rec); code != "PRODUCT_NOT_FOUND" {
		t.Errorf("got code %s", code)
	}
}

func TestCreateOrderValidationFieldsPerItem(t *testing.T) {
	body := `{"customer":" ","items":[{"sku":"A","qty":1},{"sku":"A","qty":0},{"sku":"","qty":1.5}]}`
	rec := serve(nethttp.MethodPost, "/api/orders", body)

	if rec.Code != nethttp.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	_, fields := errorCode(t, rec)
	for _, key := range []string{"customer", "items[1].sku", "items[1].qty", "items[2].sku", "items[2].qty"} {
		if fields[key] == "" {
			t.Errorf("missing field error for %s: %v", key, fields)
		}
	}
}

func TestAdjustStockRejectsBadBody(t *testing.T) {
	rec := serve(nethttp.MethodPost, "/api/products/1/stock", `{"type":"sideways","qty":0,"note":5}`)
	if rec.Code != nethttp.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	_, fields := errorCode(t, rec)
	if fields["note"] != "must be a string" {
		t.Errorf("unexpected fields: %v", fields)
	}
}

func TestCreateOrderRejectsBlankIdempotencyKey(t *testing.T) {
	rec := serveWithHeaders(
		nethttp.MethodPost,
		"/api/orders",
		`{"customer":"Budi","items":[{"sku":"A","qty":1}]}`,
		map[string]string{"Idempotency-Key": "   "},
	)

	if rec.Code != nethttp.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	code, fields := errorCode(t, rec)
	if code != "VALIDATION_ERROR" || fields["Idempotency-Key"] == "" {
		t.Errorf("unexpected error %s %v", code, fields)
	}
}
