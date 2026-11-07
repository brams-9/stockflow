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

	fields := map[string]string{}
	if sku == "" {
		fields["sku"] = "required"
	}
	if name == "" {
		fields["name"] = "required"
	}
	switch {
	case input.Price == nil:
		fields["price"] = "required"
	case *input.Price < 0:
		fields["price"] = "must be a non-negative integer"
	case *input.Price > math.MaxInt32:
		fields["price"] = "must not exceed 2147483647"
	}
	if len(fields) > 0 {
		return domain.Product{}, &domain.ValidationError{Message: "validation failed", Fields: fields}
	}

	return s.repo.Create(ctx, domain.Product{SKU: sku, Name: name, Price: *input.Price})
}
