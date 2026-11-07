package dto

type ErrorBody struct {
	Code      string            `json:"code"`
	Message   string            `json:"message"`
	Fields    map[string]string `json:"fields,omitempty"`
	SKU       string            `json:"sku,omitempty"`
	Available *int              `json:"available,omitempty"`
}

type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

func NewErrorResponse(code, message string, fields map[string]string) ErrorResponse {
	return ErrorResponse{Error: ErrorBody{Code: code, Message: message, Fields: fields}}
}
