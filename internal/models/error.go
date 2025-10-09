package models

import (
	"net/http"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    string `json:"code,omitempty"`
}

type APIError struct {
	Code    int
	Message string
	Err     error
}

func (e *APIError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

// Standard error constructors
func NewBadRequestError(message string) *APIError {
	return &APIError{Code: http.StatusBadRequest, Message: message}
}

func NewUnauthorizedError(message string) *APIError {
	return &APIError{Code: http.StatusUnauthorized, Message: message}
}

func NewNotFoundError(message string) *APIError {
	return &APIError{Code: http.StatusNotFound, Message: message}
}

func NewInternalError(message string) *APIError {
	return &APIError{Code: http.StatusInternalServerError, Message: message}
}