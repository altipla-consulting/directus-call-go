package callgo

import (
	"fmt"
	"net/http"
)

type callGoError struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Status     int               `json:"status"`
	Extensions map[string]string `json:"extensions"`
}

func (e *callGoError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func NewFailedValidationError(collection, field, message string) error {
	return &callGoError{
		Code:    "FAILED_VALIDATION",
		Message: "Validation failed",
		Status:  http.StatusBadRequest,
		Extensions: map[string]string{
			"type":       "callgo",
			"collection": collection,
			"field":      field,
			"message":    message,
		},
	}
}

func NewInvalidError(message string) error {
	return &callGoError{
		Code:    "INVALID",
		Message: message,
		Status:  http.StatusBadRequest,
		Extensions: map[string]string{
			"type": "callgo",
		},
	}
}
