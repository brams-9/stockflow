package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"stockflow/internal/domain"
	"stockflow/internal/repository"
)

func TestAdjustValidation(t *testing.T) {
	tests := []struct {
		name       string
		input      AdjustStockInput
		wantFields map[string]string
	}{
		{
			name:       "bad type",
			input:      AdjustStockInput{Type: "sideways", Qty: intPtr(1)},
			wantFields: map[string]string{"type": "must be 'in' or 'out'"},
		},
		{
			name:       "missing qty",
			input:      AdjustStockInput{Type: "in"},
			wantFields: map[string]string{"qty": "required"},
		},
		{
			name:       "zero qty",
			input:      AdjustStockInput{Type: "out", Qty: intPtr(0)},
			wantFields: map[string]string{"qty": "must be a positive integer"},
		},
		{
			name:       "qty too large",
			input:      AdjustStockInput{Type: "in", Qty: intPtr(2147483648)},
			wantFields: map[string]string{"qty": "must not exceed 2147483647"},
		},
		{
			name:  "type and qty invalid",
			input: AdjustStockInput{Type: "", Qty: intPtr(-2)},
			wantFields: map[string]string{
				"type": "must be 'in' or 'out'",
				"qty":  "must be a positive integer",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewStockService(repository.Repositories{}, nil)

			_, err := svc.Adjust(context.Background(), 1, tt.input)

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
