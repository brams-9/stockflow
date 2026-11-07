package dto

import (
	"time"

	"stockflow/internal/domain"
)

type ProductResponse struct {
	ID             int     `json:"id"`
	SKU            string  `json:"sku"`
	Name           string  `json:"name"`
	Price          int     `json:"price"`
	Stock          int     `json:"stock"`
	IsActive       bool    `json:"is_active"`
	CreatedAt      string  `json:"created_at"`
	LastMovementAt *string `json:"last_movement_at"`
}

type ProductListResponse struct {
	Data  []ProductResponse `json:"data"`
	Page  int               `json:"page"`
	Limit int               `json:"limit"`
	Total int               `json:"total"`
}

type CreateProductRequest struct {
	SKU   string `json:"sku"`
	Name  string `json:"name"`
	Price *int   `json:"price"`
}

func NewProductResponse(p domain.Product) ProductResponse {
	resp := ProductResponse{
		ID:        p.ID,
		SKU:       p.SKU,
		Name:      p.Name,
		Price:     p.Price,
		Stock:     p.Stock,
		IsActive:  p.IsActive,
		CreatedAt: formatTime(p.CreatedAt),
	}
	if p.LastMovementAt != nil {
		formatted := formatTime(*p.LastMovementAt)
		resp.LastMovementAt = &formatted
	}
	return resp
}

func NewProductListResponse(products []domain.Product, page, limit, total int) ProductListResponse {
	data := make([]ProductResponse, 0, len(products))
	for _, p := range products {
		data = append(data, NewProductResponse(p))
	}
	return ProductListResponse{Data: data, Page: page, Limit: limit, Total: total}
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}
