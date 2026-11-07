package dto

import (
	"encoding/json"

	"stockflow/internal/domain"
)

type CreateOrderItemRequest struct {
	SKU string          `json:"sku"`
	Qty json.RawMessage `json:"qty"`
}

type CreateOrderRequest struct {
	Customer string                   `json:"customer"`
	Items    []CreateOrderItemRequest `json:"items"`
}

type OrderItemResponse struct {
	ID        int    `json:"id"`
	SKU       string `json:"sku"`
	Qty       int    `json:"qty"`
	UnitPrice int    `json:"unit_price"`
}

type OrderResponse struct {
	ID          int                 `json:"id"`
	Customer    string              `json:"customer"`
	Status      string              `json:"status"`
	Total       int                 `json:"total"`
	CreatedAt   string              `json:"created_at"`
	CancelledAt *string             `json:"cancelled_at"`
	Items       []OrderItemResponse `json:"items"`
}

func NewOrderResponse(o domain.SalesOrder) OrderResponse {
	items := make([]OrderItemResponse, 0, len(o.Items))
	for _, item := range o.Items {
		items = append(items, OrderItemResponse{
			ID:        item.ID,
			SKU:       item.SKU,
			Qty:       item.Qty,
			UnitPrice: item.UnitPrice,
		})
	}

	resp := OrderResponse{
		ID:        o.ID,
		Customer:  o.Customer,
		Status:    string(o.Status),
		Total:     o.Total,
		CreatedAt: formatTime(o.CreatedAt),
		Items:     items,
	}
	if o.CancelledAt != nil {
		formatted := formatTime(*o.CancelledAt)
		resp.CancelledAt = &formatted
	}
	return resp
}
