package handler

import (
	"encoding/json"
	"errors"
	"reflect"
	"strconv"

	"github.com/labstack/echo/v4"

	"stockflow/internal/domain"
)

func parseID(c echo.Context) (int, error) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		return 0, &domain.ValidationError{
			Message: "validation failed",
			Fields:  map[string]string{"id": "must be a positive integer"},
		}
	}
	return id, nil
}

func decodeJSON(c echo.Context, dst any) error {
	if err := json.NewDecoder(c.Request().Body).Decode(dst); err != nil {
		return decodeError(err)
	}
	return nil
}

func decodeError(err error) error {
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		return &domain.ValidationError{
			Message: "validation failed",
			Fields:  map[string]string{typeErr.Field: typeMessage(typeErr)},
		}
	}
	return &domain.ValidationError{Message: "request body must be valid JSON"}
}

func typeMessage(typeErr *json.UnmarshalTypeError) string {
	switch typeErr.Type.Kind() {
	case reflect.String:
		return "must be a string"
	case reflect.Int, reflect.Int32, reflect.Int64:
		if typeErr.Field == "price" {
			return "must be a non-negative integer"
		}
		return "must be an integer"
	case reflect.Bool:
		return "must be a boolean"
	case reflect.Slice:
		return "must be an array"
	case reflect.Struct, reflect.Map:
		return "must be an object"
	default:
		return "invalid value"
	}
}
