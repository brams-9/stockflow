package handler

import (
	nethttp "net/http"

	"github.com/labstack/echo/v4"

	"stockflow/internal/delivery/http/dto"
	"stockflow/internal/service"
)

type StockHandler struct {
	service *service.StockService
}

func NewStockHandler(service *service.StockService) *StockHandler {
	return &StockHandler{service: service}
}

func (h *StockHandler) Adjust(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	var req dto.AdjustStockRequest
	if err := decodeJSON(c, &req); err != nil {
		return err
	}

	movement, err := h.service.Adjust(c.Request().Context(), id, service.AdjustStockInput{
		Type: req.Type,
		Qty:  req.Qty,
		Note: req.Note,
	})
	if err != nil {
		return err
	}

	return c.JSON(nethttp.StatusCreated, dto.NewMovementResponse(movement))
}

func (h *StockHandler) Movements(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	movements, err := h.service.Movements(c.Request().Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(nethttp.StatusOK, dto.NewMovementListResponse(movements))
}
