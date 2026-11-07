package http

import (
	"github.com/labstack/echo/v4"

	"stockflow/internal/delivery/http/handler"
)

func NewRouter(products *handler.ProductHandler) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = errorHandler

	health := NewHealthHandler()
	e.GET("/health", health.Check)

	api := e.Group("/api")
	api.GET("/products", products.List)
	api.POST("/products", products.Create)

	return e
}
