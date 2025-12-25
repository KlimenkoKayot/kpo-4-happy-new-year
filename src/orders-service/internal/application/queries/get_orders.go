package queries

import (
	"context"

	"github.com/kayotklimenko/gozon-go/orders-service/internal/domain/order"
)

type GetOrdersQuery struct {
	UserID string
}

type GetOrdersHandler struct {
	orderRepo order.Repository
}

func NewGetOrdersHandler(orderRepo order.Repository) *GetOrdersHandler {
	return &GetOrdersHandler{orderRepo: orderRepo}
}

func (h *GetOrdersHandler) Handle(ctx context.Context, query GetOrdersQuery) ([]OrderDTO, error) {
	orders, err := h.orderRepo.FindByUserID(ctx, order.UserID(query.UserID))
	if err != nil {
		return nil, err
	}

	result := make([]OrderDTO, len(orders))
	for i, o := range orders {
		result[i] = OrderDTO{
			ID:          string(o.ID()),
			UserID:      string(o.UserID()),
			Amount:      float64(o.Amount()),
			Description: o.Description(),
			Status:      string(o.Status()),
			CreatedAt:   o.CreatedAt(),
			UpdatedAt:   o.UpdatedAt(),
		}
	}

	return result, nil
}
