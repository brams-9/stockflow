package http

import "github.com/labstack/echo/v4"

func NewRouter() *echo.Echo {
	e := echo.New()
	e.HideBanner = true

	health := NewHealthHandler()
	e.GET("/health", health.Check)

	return e
}
