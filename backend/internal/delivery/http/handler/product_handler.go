package handler

import (
	nethttp "net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"stockflow/internal/delivery/http/dto"
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
	if err := decodeJSON(c, &req); err != nil {
		return err
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

func (h *ProductHandler) Update(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	var req dto.UpdateProductRequest
	if err := decodeJSON(c, &req); err != nil {
		return err
	}

	product, err := h.service.Update(c.Request().Context(), id, service.UpdateProductInput{
		SKU:      req.SKU,
		Name:     req.Name,
		Price:    req.Price,
		IsActive: req.IsActive,
	})
	if err != nil {
		return err
	}

	return c.JSON(nethttp.StatusOK, dto.NewProductResponse(product))
}

func (h *ProductHandler) Delete(c echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	if err := h.service.Delete(c.Request().Context(), id); err != nil {
		return err
	}

	return c.NoContent(nethttp.StatusNoContent)
}
