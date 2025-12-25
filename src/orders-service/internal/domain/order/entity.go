package order

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Value Objects
type OrderID string
type UserID string
type Money float64
type Status string

const (
	StatusNew       Status = "New"
	StatusFinished  Status = "Finished"
	StatusCancelled Status = "Cancelled"
)

// Errors
var (
	ErrInvalidAmount      = errors.New("amount must be greater than 0")
	ErrInvalidDescription = errors.New("description is required")
	ErrInvalidUserID      = errors.New("user ID is required")
	ErrOrderNotFound      = errors.New("order not found")
	ErrInvalidTransition  = errors.New("invalid status transition")
)

// Order - Aggregate Root
type Order struct {
	id          OrderID
	userID      UserID
	amount      Money
	description string
	status      Status
	createdAt   time.Time
	updatedAt   *time.Time
	events      []DomainEvent
}

// Factory method
func NewOrder(userID string, amount float64, description string) (*Order, error) {
	if userID == "" {
		return nil, ErrInvalidUserID
	}
	if amount <= 0 {
		return nil, ErrInvalidAmount
	}
	if description == "" {
		return nil, ErrInvalidDescription
	}

	order := &Order{
		id:          OrderID(uuid.New().String()),
		userID:      UserID(userID),
		amount:      Money(amount),
		description: description,
		status:      StatusNew,
		createdAt:   time.Now().UTC(),
		events:      make([]DomainEvent, 0),
	}

	order.addEvent(OrderCreatedEvent{
		OrderID:   string(order.id),
		UserID:    userID,
		Amount:    amount,
		CreatedAt: order.createdAt,
	})

	return order, nil
}

// Reconstruct from persistence
func ReconstructOrder(id, userID string, amount float64, description string, status Status, createdAt time.Time, updatedAt *time.Time) *Order {
	return &Order{
		id:          OrderID(id),
		userID:      UserID(userID),
		amount:      Money(amount),
		description: description,
		status:      status,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		events:      make([]DomainEvent, 0),
	}
}

// Business Logic
func (o *Order) MarkAsFinished() error {
	if o.status != StatusNew {
		return ErrInvalidTransition
	}
	o.status = StatusFinished
	now := time.Now().UTC()
	o.updatedAt = &now
	return nil
}

func (o *Order) MarkAsCancelled() error {
	if o.status != StatusNew {
		return ErrInvalidTransition
	}
	o.status = StatusCancelled
	now := time.Now().UTC()
	o.updatedAt = &now
	return nil
}

// Getters
func (o *Order) ID() OrderID           { return o.id }
func (o *Order) UserID() UserID        { return o.userID }
func (o *Order) Amount() Money         { return o.amount }
func (o *Order) Description() string   { return o.description }
func (o *Order) Status() Status        { return o.status }
func (o *Order) CreatedAt() time.Time  { return o.createdAt }
func (o *Order) UpdatedAt() *time.Time { return o.updatedAt }

// Domain Events
func (o *Order) addEvent(event DomainEvent) {
	o.events = append(o.events, event)
}

func (o *Order) PullEvents() []DomainEvent {
	events := o.events
	o.events = make([]DomainEvent, 0)
	return events
}
