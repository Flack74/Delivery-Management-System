package models

import (
	"gorm.io/gorm"
	"time"
)

type Order struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	CustomerID  uint           `gorm:"not null;index" json:"customer_id"`
	Status      OrderStatus    `gorm:"default:'created';index" json:"status"`
	Items       string         `gorm:"type:text;not null" json:"items" validate:"required"`
	Description string         `gorm:"type:text" json:"description"`
	Address     string         `gorm:"type:text;not null" json:"address" validate:"required"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Customer    User           `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
}

type CreateOrderRequest struct {
	Items       string `json:"items" validate:"required"`
	Description string `json:"description"`
	Address     string `json:"address" validate:"required"`
}

type UpdateOrderStatusRequest struct {
	Status OrderStatus `json:"status" validate:"required"`
}

type OrderResponse struct {
	ID          uint          `json:"id"`
	CustomerID  uint          `json:"customer_id"`
	Status      OrderStatus   `json:"status"`
	Items       string        `json:"items"`
	Description string        `json:"description"`
	Address     string        `json:"address"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
	Customer    *UserResponse `json:"customer,omitempty"`
}

func (o *Order) ToResponse() *OrderResponse {
	resp := &OrderResponse{
		ID:          o.ID,
		CustomerID:  o.CustomerID,
		Status:      o.Status,
		Items:       o.Items,
		Description: o.Description,
		Address:     o.Address,
		CreatedAt:   o.CreatedAt,
		UpdatedAt:   o.UpdatedAt,
	}
	if o.Customer.ID != 0 {
		resp.Customer = o.Customer.ToResponse()
	}
	return resp
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code,omitempty"`
}

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
	Type    string `json:"type"`
}

func NewAPIError(code int, message, details, errorType string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Details: details,
		Type:    errorType,
	}
}

func (e *APIError) Error() string {
	return e.Message
}
