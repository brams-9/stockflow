package handler

import (
	nethttp "net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"stockflow/internal/delivery/http/dto"
	"stockflow/internal/domain"
	"stockflow/internal/service"
)

type OrderHandler struct {
	service *service.OrderService
}

func NewOrderHandler(service *service.OrderService) *OrderHandler {
	return &OrderHandler{service: service}
}

func (h *OrderHandler) Create(c echo.Context) error {
	var req dto.CreateOrderRequest
	if err := decodeJSON(c, &req); err != nil {
		return err
	}

	items := make([]service.OrderItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		qty, _ := strconv.Atoi(string(item.Qty))
		items = append(items, service.OrderItemInput{SKU: item.SKU, Qty: qty})
	}

	input := service.CreateOrderInput{Customer: req.Customer, Items: items}

	var order domain.SalesOrder
	var replayed bool
	var err error
	if keys := c.Request().Header.Values("Idempotency-Key"); len(keys) > 0 {
		order, replayed, err = h.service.CreateIdempotent(c.Request().Context(), keys[0], input)
	} else {
		order, err = h.service.Create(c.Request().Context(), input)
	}
	if err != nil {
		return err
	}

	if replayed {
		c.Response().Header().Set("Idempotent-Replay", "true")
	}
	return c.JSON(nethttp.StatusCreated, dto.NewOrderResponse(order))
}

func (h *OrderHandler) Get(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	order, err := h.service.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(nethttp.StatusOK, dto.NewOrderResponse(order))
}

func (h *OrderHandler) Cancel(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	order, err := h.service.Cancel(c.Request().Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(nethttp.StatusOK, dto.NewOrderResponse(order))
}
