package commands

import (
	"context"
	"log"

	"github.com/kayotklimenko/gozon-go/orders-service/internal/domain/order"
)

type UpdateOrderStatusCommand struct {
	OrderID string
	Success bool
}

type UpdateOrderStatusHandler struct {
	orderRepo order.Repository
}

func NewUpdateOrderStatusHandler(orderRepo order.Repository) *UpdateOrderStatusHandler {
	return &UpdateOrderStatusHandler{
		orderRepo: orderRepo,
	}
}

func (h *UpdateOrderStatusHandler) Handle(ctx context.Context, cmd UpdateOrderStatusCommand) error {
	o, err := h.orderRepo.FindByIDOnly(ctx, order.OrderID(cmd.OrderID))
	if err != nil {
		log.Printf("Order %s not found: %v", cmd.OrderID, err)
		return nil
	}

	if cmd.Success {
		if err := o.MarkAsFinished(); err != nil {
			log.Printf("Order %s already has final status", cmd.OrderID)
			return nil
		}
	} else {
		if err := o.MarkAsCancelled(); err != nil {
			log.Printf("Order %s already has final status", cmd.OrderID)
			return nil
		}
	}

	if err := h.orderRepo.Update(ctx, o); err != nil {
		log.Printf("Failed to update order %s: %v", cmd.OrderID, err)
		return err
	}

	log.Printf("Order %s status updated to %s", cmd.OrderID, o.Status())
	return nil
}
