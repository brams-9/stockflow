package http

import (
	nethttp "net/http"

	"github.com/labstack/echo/v4"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Check(c echo.Context) error {
	return c.JSON(nethttp.StatusOK, map[string]string{"status": "ok"})
}
