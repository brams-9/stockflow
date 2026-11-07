package http

import (
	"github.com/labstack/echo/v4"

	"stockflow/internal/delivery/http/handler"
)

func NewRouter(products *handler.ProductHandler, stock *handler.StockHandler, orders *handler.OrderHandler) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HTTPErrorHandler = errorHandler

	health := NewHealthHandler()
	e.GET("/health", health.Check)

	api := e.Group("/api")
	api.GET("/products", products.List)
	api.POST("/products", products.Create)
	api.PUT("/products/:id", products.Update)
	api.DELETE("/products/:id", products.Delete)
	api.POST("/products/:id/stock", stock.Adjust)
	api.GET("/products/:id/movements", stock.Movements)

	api.POST("/orders", orders.Create)
	api.GET("/orders/:id", orders.Get)
	api.POST("/orders/:id/cancel", orders.Cancel)

	return e
}
