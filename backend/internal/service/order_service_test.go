package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"stockflow/internal/domain"
	"stockflow/internal/repository"
)

func TestCreateOrderValidation(t *testing.T) {
	tests := []struct {
		name       string
		input      CreateOrderInput
		wantFields map[string]string
	}{
		{
			name:  "blank customer and no items",
			input: CreateOrderInput{Customer: "  "},
			wantFields: map[string]string{
				"customer": "required",
				"items":    "at least one item is required",
			},
		},
		{
			name: "per item errors",
			input: CreateOrderInput{
				Customer: "Budi",
				Items: []OrderItemInput{
					{SKU: "A", Qty: 1},
					{SKU: "", Qty: 0},
					{SKU: "A", Qty: -3},
				},
			},
			wantFields: map[string]string{
				"items[1].sku": "required",
				"items[1].qty": "must be a positive integer",
				"items[2].sku": "duplicate sku",
				"items[2].qty": "must be a positive integer",
			},
		},
		{
			name: "qty too large",
			input: CreateOrderInput{
				Customer: "Budi",
				Items:    []OrderItemInput{{SKU: "A", Qty: 2147483648}},
			},
			wantFields: map[string]string{"items[0].qty": "must not exceed 2147483647"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewOrderService(repository.Repositories{}, nil)

			_, err := svc.Create(context.Background(), tt.input)

			var validationErr *domain.ValidationError
			if !errors.As(err, &validationErr) {
				t.Fatalf("expected validation error, got %v", err)
			}
			if !reflect.DeepEqual(validationErr.Fields, tt.wantFields) {
				t.Errorf("got fields %v, want %v", validationErr.Fields, tt.wantFields)
			}
		})
	}
}
