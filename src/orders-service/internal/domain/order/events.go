package order

import "time"

type DomainEvent interface {
	EventName() string
}

type OrderCreatedEvent struct {
	OrderID   string
	UserID    string
	Amount    float64
	CreatedAt time.Time
}

func (e OrderCreatedEvent) EventName() string {
	return "OrderCreated"
}

type OrderStatusChangedEvent struct {
	OrderID   string
	OldStatus Status
	NewStatus Status
	ChangedAt time.Time
}

func (e OrderStatusChangedEvent) EventName() string {
	return "OrderStatusChanged"
}
