package commands

import (
	"context"
	"encoding/json"
	"time"

	"github.com/kayotklimenko/gozon-go/orders-service/internal/domain/order"
	"github.com/kayotklimenko/gozon-go/orders-service/internal/domain/outbox"
)

type CreateOrderCommand struct {
	UserID      string
	Amount      float64
	Description string
}

type CreateOrderResult struct {
	OrderID     string  `json:"id"`
	UserID      string  `json:"user_id"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
}

type PaymentRequestMessage struct {
	OrderID        string    `json:"order_id"`
	UserID         string    `json:"user_id"`
	Amount         float64   `json:"amount"`
	CreatedAt      time.Time `json:"created_at"`
	IdempotencyKey string    `json:"idempotency_key"`
}

type UnitOfWork interface {
	Begin(ctx context.Context) (context.Context, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type CreateOrderHandler struct {
	orderRepo  order.Repository
	outboxRepo outbox.Repository
	unitOfWork UnitOfWork
}

func NewCreateOrderHandler(orderRepo order.Repository, outboxRepo outbox.Repository, uow UnitOfWork) *CreateOrderHandler {
	return &CreateOrderHandler{
		orderRepo:  orderRepo,
		outboxRepo: outboxRepo,
		unitOfWork: uow,
	}
}

func (h *CreateOrderHandler) Handle(ctx context.Context, cmd CreateOrderCommand) (*CreateOrderResult, error) {
	newOrder, err := order.NewOrder(cmd.UserID, cmd.Amount, cmd.Description)
	if err != nil {
		return nil, err
	}

	paymentRequest := PaymentRequestMessage{
		OrderID:        string(newOrder.ID()),
		UserID:         cmd.UserID,
		Amount:         cmd.Amount,
		CreatedAt:      time.Now().UTC(),
		IdempotencyKey: string(newOrder.ID()),
	}

	payload, err := json.Marshal(paymentRequest)
	if err != nil {
		return nil, err
	}

	outboxMsg := outbox.NewMessage("PaymentRequestMessage", string(payload))

	txCtx, err := h.unitOfWork.Begin(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.orderRepo.Save(txCtx, newOrder); err != nil {
		h.unitOfWork.Rollback(txCtx)
		return nil, err
	}

	if err := h.outboxRepo.Save(txCtx, outboxMsg); err != nil {
		h.unitOfWork.Rollback(txCtx)
		return nil, err
	}

	if err := h.unitOfWork.Commit(txCtx); err != nil {
		return nil, err
	}

	return &CreateOrderResult{
		OrderID:     string(newOrder.ID()),
		UserID:      string(newOrder.UserID()),
		Amount:      float64(newOrder.Amount()),
		Description: newOrder.Description(),
		Status:      string(newOrder.Status()),
		CreatedAt:   newOrder.CreatedAt().Format(time.RFC3339),
	}, nil
}
