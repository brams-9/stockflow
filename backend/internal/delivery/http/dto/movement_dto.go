package dto

import "stockflow/internal/domain"

type AdjustStockRequest struct {
	Type string  `json:"type"`
	Qty  *int    `json:"qty"`
	Note *string `json:"note"`
}

type MovementResponse struct {
	ID        int     `json:"id"`
	ProductID int     `json:"product_id"`
	Type      string  `json:"type"`
	Qty       int     `json:"qty"`
	Note      *string `json:"note"`
	CreatedAt string  `json:"created_at"`
}

type MovementListResponse struct {
	Data []MovementResponse `json:"data"`
}

func NewMovementResponse(m domain.StockMovement) MovementResponse {
	return MovementResponse{
		ID:        m.ID,
		ProductID: m.ProductID,
		Type:      string(m.Type),
		Qty:       m.Qty,
		Note:      m.Note,
		CreatedAt: formatTime(m.CreatedAt),
	}
}

func NewMovementListResponse(movements []domain.StockMovement) MovementListResponse {
	data := make([]MovementResponse, 0, len(movements))
	for _, m := range movements {
		data = append(data, NewMovementResponse(m))
	}
	return MovementListResponse{Data: data}
}
