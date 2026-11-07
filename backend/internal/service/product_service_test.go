package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"stockflow/internal/domain"
)

type fakeProductRepository struct {
	updated    domain.Product
	deletedID  int
	listSearch string
	listLimit  int
	listOffset int
	created    domain.Product
	createErr  error
	called     bool
}

func (f *fakeProductRepository) List(ctx context.Context, search string, limit, offset int) ([]domain.Product, int, error) {
	f.listSearch = search
	f.listLimit = limit
	f.listOffset = offset
	return []domain.Product{}, 42, nil
}

func (f *fakeProductRepository) Create(ctx context.Context, product domain.Product) (domain.Product, error) {
	f.called = true
	f.created = product
	if f.createErr != nil {
		return domain.Product{}, f.createErr
	}
	product.ID = 1
	product.IsActive = true
	return product, nil
}

func (f *fakeProductRepository) GetByID(ctx context.Context, id int) (domain.Product, error) {
	return domain.Product{ID: id}, nil
}

func (f *fakeProductRepository) FindBySKUs(ctx context.Context, skus []string) ([]domain.Product, error) {
	return []domain.Product{}, nil
}

func (f *fakeProductRepository) Update(ctx context.Context, product domain.Product) (domain.Product, error) {
	f.called = true
	f.updated = product
	return product, nil
}

func (f *fakeProductRepository) Delete(ctx context.Context, id int) error {
	f.deletedID = id
	return nil
}

func intPtr(v int) *int {
	return &v
}

func TestListNormalizesPaging(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		limit      int
		wantPage   int
		wantLimit  int
		wantOffset int
	}{
		{"defaults when zero", 0, 0, 1, 10, 0},
		{"defaults when negative", -3, -1, 1, 10, 0},
		{"second page", 2, 5, 2, 5, 5},
		{"limit capped", 1, 1000, 1, 100, 0},
		{"offset follows capped limit", 3, 500, 3, 100, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeProductRepository{}
			svc := NewProductService(repo)

			result, err := svc.List(context.Background(), "", tt.page, tt.limit)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Page != tt.wantPage || result.Limit != tt.wantLimit {
				t.Errorf("got page=%d limit=%d, want page=%d limit=%d", result.Page, result.Limit, tt.wantPage, tt.wantLimit)
			}
			if repo.listLimit != tt.wantLimit || repo.listOffset != tt.wantOffset {
				t.Errorf("repo got limit=%d offset=%d, want limit=%d offset=%d", repo.listLimit, repo.listOffset, tt.wantLimit, tt.wantOffset)
			}
			if result.Total != 42 {
				t.Errorf("got total=%d, want 42", result.Total)
			}
		})
	}
}

func TestListTrimsSearch(t *testing.T) {
	repo := &fakeProductRepository{}
	svc := NewProductService(repo)

	if _, err := svc.List(context.Background(), "  rice ", 1, 10); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.listSearch != "rice" {
		t.Errorf("got search %q, want %q", repo.listSearch, "rice")
	}
}

func TestCreateValidation(t *testing.T) {
	tests := []struct {
		name       string
		input      CreateProductInput
		wantFields map[string]string
	}{
		{
			name:       "missing sku",
			input:      CreateProductInput{Name: "Rice", Price: intPtr(100)},
			wantFields: map[string]string{"sku": "required"},
		},
		{
			name:       "blank sku",
			input:      CreateProductInput{SKU: "   ", Name: "Rice", Price: intPtr(100)},
			wantFields: map[string]string{"sku": "required"},
		},
		{
			name:       "blank name",
			input:      CreateProductInput{SKU: "RC-1", Name: "  ", Price: intPtr(100)},
			wantFields: map[string]string{"name": "required"},
		},
		{
			name:       "missing price",
			input:      CreateProductInput{SKU: "RC-1", Name: "Rice"},
			wantFields: map[string]string{"price": "required"},
		},
		{
			name:       "negative price",
			input:      CreateProductInput{SKU: "RC-1", Name: "Rice", Price: intPtr(-1)},
			wantFields: map[string]string{"price": "must be a non-negative integer"},
		},
		{
			name:       "price too large",
			input:      CreateProductInput{SKU: "RC-1", Name: "Rice", Price: intPtr(2147483648)},
			wantFields: map[string]string{"price": "must not exceed 2147483647"},
		},
		{
			name:  "all fields invalid",
			input: CreateProductInput{SKU: " ", Name: "", Price: intPtr(-5)},
			wantFields: map[string]string{
				"sku":   "required",
				"name":  "required",
				"price": "must be a non-negative integer",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeProductRepository{}
			svc := NewProductService(repo)

			_, err := svc.Create(context.Background(), tt.input)

			var validationErr *domain.ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("expected validation error, got %v", err)
			}
			if !reflect.DeepEqual(validationErr.Fields, tt.wantFields) {
				t.Errorf("got fields %v, want %v", validationErr.Fields, tt.wantFields)
			}
			if repo.called {
				t.Error("repository must not be called when validation fails")
			}
		})
	}
}

func TestCreateAcceptsZeroPriceAndTrims(t *testing.T) {
	repo := &fakeProductRepository{}
	svc := NewProductService(repo)

	product, err := svc.Create(context.Background(), CreateProductInput{SKU: "  RC-1 ", Name: " Rice ", Price: intPtr(0)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.created.SKU != "RC-1" || repo.created.Name != "Rice" || repo.created.Price != 0 {
		t.Errorf("repository received %+v", repo.created)
	}
	if product.ID != 1 {
		t.Errorf("got id %d, want 1", product.ID)
	}
}

func TestCreatePropagatesDuplicateSKU(t *testing.T) {
	repo := &fakeProductRepository{createErr: domain.ErrSKUExists}
	svc := NewProductService(repo)

	_, err := svc.Create(context.Background(), CreateProductInput{SKU: "RC-1", Name: "Rice", Price: intPtr(10)})
	if !errors.Is(err, domain.ErrSKUExists) {
		t.Fatalf("expected ErrSKUExists, got %v", err)
	}
}

func boolPtr(v bool) *bool {
	return &v
}

func TestUpdateValidation(t *testing.T) {
	tests := []struct {
		name       string
		input      UpdateProductInput
		wantFields map[string]string
	}{
		{
			name:       "missing is_active",
			input:      UpdateProductInput{SKU: "RC-1", Name: "Rice", Price: intPtr(10)},
			wantFields: map[string]string{"is_active": "required"},
		},
		{
			name:  "everything invalid",
			input: UpdateProductInput{SKU: "", Name: " ", Price: intPtr(-1)},
			wantFields: map[string]string{
				"sku":       "required",
				"name":      "required",
				"price":     "must be a non-negative integer",
				"is_active": "required",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeProductRepository{}
			svc := NewProductService(repo)

			_, err := svc.Update(context.Background(), 1, tt.input)

			var validationErr *domain.ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("expected validation error, got %v", err)
			}
			if !reflect.DeepEqual(validationErr.Fields, tt.wantFields) {
				t.Errorf("got fields %v, want %v", validationErr.Fields, tt.wantFields)
			}
			if repo.called {
				t.Error("repository must not be called when validation fails")
			}
		})
	}
}

func TestUpdateTrimsAndPassesID(t *testing.T) {
	repo := &fakeProductRepository{}
	svc := NewProductService(repo)

	_, err := svc.Update(context.Background(), 9, UpdateProductInput{SKU: " RC-1 ", Name: " Rice ", Price: intPtr(0), IsActive: boolPtr(false)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := domain.Product{ID: 9, SKU: "RC-1", Name: "Rice", Price: 0, IsActive: false}
	if repo.updated != want {
		t.Errorf("repository received %+v, want %+v", repo.updated, want)
	}
}
