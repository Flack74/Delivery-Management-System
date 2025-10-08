package models

type OrderStatus string

const (
	StatusCreated    OrderStatus = "created"
	StatusDispatched OrderStatus = "dispatched"
	StatusInTransit  OrderStatus = "in_transit"
	StatusDelivered  OrderStatus = "delivered"
	StatusCancelled  OrderStatus = "cancelled"
)

var statusTransitions = map[OrderStatus][]OrderStatus{
	StatusCreated:    {StatusDispatched, StatusCancelled},
	StatusDispatched: {StatusInTransit, StatusCancelled},
	StatusInTransit:  {StatusDelivered},
	StatusDelivered:  {},
	StatusCancelled:  {},
}

func (s OrderStatus) CanTransitionTo(newStatus OrderStatus) bool {
	allowedTransitions, exists := statusTransitions[s]
	if !exists {
		return false
	}
	
	for _, allowed := range allowedTransitions {
		if allowed == newStatus {
			return true
		}
	}
	return false
}

func (s OrderStatus) IsValid() bool {
	switch s {
	case StatusCreated, StatusDispatched, StatusInTransit, StatusDelivered, StatusCancelled:
		return true
	default:
		return false
	}
}

func (s OrderStatus) IsFinal() bool {
	return s == StatusDelivered || s == StatusCancelled
}