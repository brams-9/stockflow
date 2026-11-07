package domain

import "time"

type OrderStatus string

const (
	OrderCreated   OrderStatus = "CREATED"
	OrderCancelled OrderStatus = "CANCELLED"
)

type SalesOrder struct {
	ID          int
	Customer    string
	Status      OrderStatus
	Total       int
	CreatedAt   time.Time
	CancelledAt *time.Time
	Items       []SalesOrderItem
}

type SalesOrderItem struct {
	ID        int
	OrderID   int
	SKU       string
	Qty       int
	UnitPrice int
}
