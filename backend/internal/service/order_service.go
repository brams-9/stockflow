package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"unicode/utf8"

	"stockflow/internal/domain"
	"stockflow/internal/repository"
)

type OrderItemInput struct {
	SKU string
	Qty int
}

type CreateOrderInput struct {
	Customer string
	Items    []OrderItemInput
}

const maxIdempotencyKeyLength = 255

type orderLine struct {
	index int
	sku   string
	qty   int
}

type OrderService struct {
	repos repository.Repositories
	tx    repository.TxRunner
}

func NewOrderService(repos repository.Repositories, tx repository.TxRunner) *OrderService {
	return &OrderService{repos: repos, tx: tx}
}

func (s *OrderService) Create(ctx context.Context, input CreateOrderInput) (domain.SalesOrder, error) {
	order, _, err := s.create(ctx, nil, input)
	return order, err
}

func (s *OrderService) CreateIdempotent(ctx context.Context, key string, input CreateOrderInput) (domain.SalesOrder, bool, error) {
	return s.create(ctx, &key, input)
}

func (s *OrderService) create(ctx context.Context, key *string, input CreateOrderInput) (domain.SalesOrder, bool, error) {
	customer := strings.TrimSpace(input.Customer)

	fields := map[string]string{}
	idempotencyKey := ""
	if key != nil {
		idempotencyKey = strings.TrimSpace(*key)
		if n := utf8.RuneCountInString(idempotencyKey); n == 0 || n > maxIdempotencyKeyLength {
			fields["Idempotency-Key"] = "must be between 1 and 255 characters"
		}
	}
	if customer == "" {
		fields["customer"] = "required"
	}
	if len(input.Items) == 0 {
		fields["items"] = "at least one item is required"
	}

	seen := map[string]bool{}
	lines := make([]orderLine, 0, len(input.Items))
	for i, item := range input.Items {
		sku := strings.TrimSpace(item.SKU)
		switch {
		case sku == "":
			fields[itemField(i, "sku")] = "required"
		case seen[sku]:
			fields[itemField(i, "sku")] = "duplicate sku"
		default:
			seen[sku] = true
		}

		switch {
		case item.Qty <= 0:
			fields[itemField(i, "qty")] = "must be a positive integer"
		case item.Qty > math.MaxInt32:
			fields[itemField(i, "qty")] = "must not exceed 2147483647"
		}

		lines = append(lines, orderLine{index: i, sku: sku, qty: item.Qty})
	}
	if len(fields) > 0 {
		return domain.SalesOrder{}, false, validationFailed(fields)
	}

	sort.Slice(lines, func(a, b int) bool { return lines[a].sku < lines[b].sku })
	requestHash := fingerprint(customer, lines)

	var order domain.SalesOrder
	replayed := false
	err := s.tx.WithinTx(ctx, func(r repository.Repositories) error {
		if key != nil {
			claimed, err := r.Idempotency.Claim(ctx, idempotencyKey, requestHash)
			if err != nil {
				return err
			}
			if !claimed {
				existing, err := s.replay(ctx, r, idempotencyKey, requestHash)
				if err != nil {
					return err
				}
				order = existing
				replayed = true
				return nil
			}
		}

		placed, err := s.place(ctx, r, customer, lines)
		if err != nil {
			return err
		}
		order = placed

		if key != nil {
			return r.Idempotency.AttachOrder(ctx, idempotencyKey, order.ID)
		}
		return nil
	})
	if err != nil {
		return domain.SalesOrder{}, false, err
	}
	return order, replayed, nil
}

func (s *OrderService) replay(ctx context.Context, r repository.Repositories, key, requestHash string) (domain.SalesOrder, error) {
	record, err := r.Idempotency.Find(ctx, key)
	if err != nil {
		return domain.SalesOrder{}, err
	}
	if record.RequestHash != requestHash {
		return domain.SalesOrder{}, domain.ErrIdempotencyKeyReused
	}
	return r.Orders.GetByID(ctx, record.OrderID)
}

func fingerprint(customer string, lines []orderLine) string {
	type item struct {
		SKU string `json:"sku"`
		Qty int    `json:"qty"`
	}
	payload := struct {
		Customer string `json:"customer"`
		Items    []item `json:"items"`
	}{Customer: customer, Items: make([]item, 0, len(lines))}
	for _, line := range lines {
		payload.Items = append(payload.Items, item{SKU: line.sku, Qty: line.qty})
	}

	raw, _ := json.Marshal(payload)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func (s *OrderService) place(ctx context.Context, r repository.Repositories, customer string, lines []orderLine) (domain.SalesOrder, error) {
	skus := make([]string, 0, len(lines))
	for _, line := range lines {
		skus = append(skus, line.sku)
	}

	products, err := r.Products.FindBySKUs(ctx, skus)
	if err != nil {
		return domain.SalesOrder{}, err
	}
	bySKU := make(map[string]domain.Product, len(products))
	for _, p := range products {
		bySKU[p.SKU] = p
	}

	fields := map[string]string{}
	for _, line := range lines {
		product, found := bySKU[line.sku]
		switch {
		case !found:
			fields[itemField(line.index, "sku")] = "unknown sku"
		case !product.IsActive:
			fields[itemField(line.index, "sku")] = "product is inactive"
		}
	}
	if len(fields) > 0 {
		return domain.SalesOrder{}, validationFailed(fields)
	}

	reservations := make([]repository.Reservation, 0, len(lines))
	items := make([]domain.SalesOrderItem, 0, len(lines))
	total := 0
	for _, line := range lines {
		reservation, err := r.Stock.DecreaseBySKU(ctx, line.sku, line.qty)
		if err != nil {
			return domain.SalesOrder{}, lineError(line, err)
		}

		total += reservation.UnitPrice * line.qty
		if total > math.MaxInt32 {
			return domain.SalesOrder{}, validationFailed(map[string]string{"total": "must not exceed 2147483647"})
		}

		reservations = append(reservations, reservation)
		items = append(items, domain.SalesOrderItem{SKU: line.sku, Qty: line.qty, UnitPrice: reservation.UnitPrice})
	}

	order, err := r.Orders.Create(ctx, domain.SalesOrder{Customer: customer, Total: total, Items: items})
	if err != nil {
		return domain.SalesOrder{}, err
	}

	note := fmt.Sprintf("order #%d", order.ID)
	for i, item := range order.Items {
		_, err := r.Movements.Create(ctx, domain.StockMovement{
			ProductID: reservations[i].ProductID,
			Type:      domain.MovementOut,
			Qty:       item.Qty,
			Note:      &note,
		})
		if err != nil {
			return domain.SalesOrder{}, err
		}
	}
	return order, nil
}

func (s *OrderService) Get(ctx context.Context, id int) (domain.SalesOrder, error) {
	return s.repos.Orders.GetByID(ctx, id)
}

func (s *OrderService) Cancel(ctx context.Context, id int) (domain.SalesOrder, error) {
	var cancelled domain.SalesOrder
	err := s.tx.WithinTx(ctx, func(r repository.Repositories) error {
		order, err := r.Orders.GetByIDForUpdate(ctx, id)
		if err != nil {
			return err
		}
		if order.Status == domain.OrderCancelled {
			return domain.ErrOrderAlreadyCancelled
		}

		note := fmt.Sprintf("cancel order #%d", order.ID)
		for _, item := range order.Items {
			productID, err := r.Stock.IncreaseBySKU(ctx, item.SKU, item.Qty)
			if err != nil {
				return err
			}
			_, err = r.Movements.Create(ctx, domain.StockMovement{
				ProductID: productID,
				Type:      domain.MovementIn,
				Qty:       item.Qty,
				Note:      &note,
			})
			if err != nil {
				return err
			}
		}

		cancelled, err = r.Orders.MarkCancelled(ctx, order.ID)
		return err
	})
	if err != nil {
		return domain.SalesOrder{}, err
	}
	return cancelled, nil
}

func lineError(line orderLine, err error) error {
	switch {
	case errors.Is(err, domain.ErrProductNotFound):
		return validationFailed(map[string]string{itemField(line.index, "sku"): "unknown sku"})
	case errors.Is(err, domain.ErrProductInactive):
		return validationFailed(map[string]string{itemField(line.index, "sku"): "product is inactive"})
	default:
		return err
	}
}

func itemField(index int, name string) string {
	return fmt.Sprintf("items[%d].%s", index, name)
}

func validationFailed(fields map[string]string) error {
	return &domain.ValidationError{Message: "validation failed", Fields: fields}
}
