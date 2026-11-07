package handler

import (
	"encoding/json"
	"errors"
	nethttp "net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"stockflow/internal/delivery/http/dto"
	"stockflow/internal/domain"
	"stockflow/internal/service"
)

type ProductHandler struct {
	service *service.ProductService
}

func NewProductHandler(service *service.ProductService) *ProductHandler {
	return &ProductHandler{service: service}
}

func (h *ProductHandler) List(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	limit, _ := strconv.Atoi(c.QueryParam("limit"))

	result, err := h.service.List(c.Request().Context(), c.QueryParam("search"), page, limit)
	if err != nil {
		return err
	}

	return c.JSON(nethttp.StatusOK, dto.NewProductListResponse(result.Products, result.Page, result.Limit, result.Total))
}

func (h *ProductHandler) Create(c echo.Context) error {
	var req dto.CreateProductRequest
	if err := json.NewDecoder(c.Request().Body).Decode(&req); err != nil {
		return decodeError(err)
	}

	product, err := h.service.Create(c.Request().Context(), service.CreateProductInput{
		SKU:   req.SKU,
		Name:  req.Name,
		Price: req.Price,
	})
	if err != nil {
		return err
	}

	return c.JSON(nethttp.StatusCreated, dto.NewProductResponse(product))
}

func decodeError(err error) error {
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		message := "must be a string"
		if typeErr.Field == "price" {
			message = "must be a non-negative integer"
		}
		return &domain.ValidationError{
			Message: "validation failed",
			Fields:  map[string]string{typeErr.Field: message},
		}
	}
	return &domain.ValidationError{Message: "request body must be valid JSON"}
}
