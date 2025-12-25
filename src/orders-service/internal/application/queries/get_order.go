package queries

import (
	"context"
	"time"

	"github.com/kayotklimenko/gozon-go/orders-service/internal/domain/order"
)

type GetOrderQuery struct {
	OrderID string
	UserID  string
}

type OrderDTO struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Amount      float64    `json:"amount"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

type GetOrderHandler struct {
	orderRepo order.Repository
}

func NewGetOrderHandler(orderRepo order.Repository) *GetOrderHandler {
	return &GetOrderHandler{orderRepo: orderRepo}
}

func (h *GetOrderHandler) Handle(ctx context.Context, query GetOrderQuery) (*OrderDTO, error) {
	o, err := h.orderRepo.FindByID(ctx, order.OrderID(query.OrderID), order.UserID(query.UserID))
	if err != nil {
		return nil, err
	}

	return &OrderDTO{
		ID:          string(o.ID()),
		UserID:      string(o.UserID()),
		Amount:      float64(o.Amount()),
		Description: o.Description(),
		Status:      string(o.Status()),
		CreatedAt:   o.CreatedAt(),
		UpdatedAt:   o.UpdatedAt(),
	}, nil
}
