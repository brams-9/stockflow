package http

import (
	"errors"
	"log"
	nethttp "net/http"

	"github.com/labstack/echo/v4"

	"stockflow/internal/delivery/http/dto"
	"stockflow/internal/domain"
)

func errorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	status, body := mapError(err, c)
	if jsonErr := c.JSON(status, body); jsonErr != nil {
		log.Printf("write error response: %v", jsonErr)
	}
}

func mapError(err error, c echo.Context) (int, dto.ErrorResponse) {
	var validationErr *domain.ValidationError
	var httpErr *echo.HTTPError

	switch {
	case errors.As(err, &validationErr):
		return nethttp.StatusBadRequest, dto.NewErrorResponse("VALIDATION_ERROR", validationErr.Message, validationErr.Fields)
	case errors.Is(err, domain.ErrSKUExists):
		return nethttp.StatusConflict, dto.NewErrorResponse("SKU_ALREADY_EXISTS", "a product with this sku already exists", nil)
	case errors.As(err, &httpErr) && httpErr.Code == nethttp.StatusNotFound:
		return nethttp.StatusNotFound, dto.NewErrorResponse("NOT_FOUND", "resource not found", nil)
	case errors.As(err, &httpErr) && httpErr.Code == nethttp.StatusMethodNotAllowed:
		return nethttp.StatusMethodNotAllowed, dto.NewErrorResponse("METHOD_NOT_ALLOWED", "method not allowed", nil)
	default:
		log.Printf("%s %s: %v", c.Request().Method, c.Request().URL.Path, err)
		return nethttp.StatusInternalServerError, dto.NewErrorResponse("INTERNAL_ERROR", "internal server error", nil)
	}
}
