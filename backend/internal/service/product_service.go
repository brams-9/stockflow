package service

import (
	"context"
	"math"
	"strings"

	"stockflow/internal/domain"
	"stockflow/internal/repository"
)

const (
	defaultPage  = 1
	defaultLimit = 10
	maxLimit     = 100
)

type ProductList struct {
	Products []domain.Product
	Total    int
	Page     int
	Limit    int
}

type CreateProductInput struct {
	SKU   string
	Name  string
	Price *int
}

type UpdateProductInput struct {
	SKU      string
	Name     string
	Price    *int
	IsActive *bool
}

type ProductService struct {
	repo repository.ProductRepository
}

func NewProductService(repo repository.ProductRepository) *ProductService {
	return &ProductService{repo: repo}
}

func (s *ProductService) List(ctx context.Context, search string, page, limit int) (ProductList, error) {
	if page < 1 {
		page = defaultPage
	}
	if limit < 1 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	products, total, err := s.repo.List(ctx, strings.TrimSpace(search), limit, (page-1)*limit)
	if err != nil {
		return ProductList{}, err
	}

	return ProductList{Products: products, Total: total, Page: page, Limit: limit}, nil
}

func (s *ProductService) Create(ctx context.Context, input CreateProductInput) (domain.Product, error) {
	sku := strings.TrimSpace(input.SKU)
	name := strings.TrimSpace(input.Name)

	if fields := validateProductFields(sku, name, input.Price); len(fields) > 0 {
		return domain.Product{}, &domain.ValidationError{Message: "validation failed", Fields: fields}
	}

	return s.repo.Create(ctx, domain.Product{SKU: sku, Name: name, Price: *input.Price})
}

func (s *ProductService) Update(ctx context.Context, id int, input UpdateProductInput) (domain.Product, error) {
	sku := strings.TrimSpace(input.SKU)
	name := strings.TrimSpace(input.Name)

	fields := validateProductFields(sku, name, input.Price)
	if input.IsActive == nil {
		fields["is_active"] = "required"
	}
	if len(fields) > 0 {
		return domain.Product{}, &domain.ValidationError{Message: "validation failed", Fields: fields}
	}

	return s.repo.Update(ctx, domain.Product{
		ID:       id,
		SKU:      sku,
		Name:     name,
		Price:    *input.Price,
		IsActive: *input.IsActive,
	})
}

func (s *ProductService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func validateProductFields(sku, name string, price *int) map[string]string {
	fields := map[string]string{}
	if sku == "" {
		fields["sku"] = "required"
	}
	if name == "" {
		fields["name"] = "required"
	}
	switch {
	case price == nil:
		fields["price"] = "required"
	case *price < 0:
		fields["price"] = "must be a non-negative integer"
	case *price > math.MaxInt32:
		fields["price"] = "must not exceed 2147483647"
	}
	return fields
}
