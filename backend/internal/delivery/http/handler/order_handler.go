package handler

import (
	nethttp "net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"stockflow/internal/delivery/http/dto"
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

	order, err := h.service.Create(c.Request().Context(), service.CreateOrderInput{
		Customer: req.Customer,
		Items:    items,
	})
	if err != nil {
		return err
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
