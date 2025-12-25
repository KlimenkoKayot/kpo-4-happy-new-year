package rabbitmq

import (
	"context"
	"encoding/json"
	"log"

	"github.com/kayotklimenko/gozon-go/orders-service/internal/application/commands"
	"github.com/kayotklimenko/gozon-go/orders-service/internal/infrastructure/messaging"
	amqp "github.com/rabbitmq/amqp091-go"
)

type PaymentResultConsumer struct {
	channel             *amqp.Channel
	updateStatusHandler *commands.UpdateOrderStatusHandler
}

func NewPaymentResultConsumer(conn *amqp.Connection, updateStatusHandler *commands.UpdateOrderStatusHandler) (*PaymentResultConsumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &PaymentResultConsumer{
		channel:             ch,
		updateStatusHandler: updateStatusHandler,
	}, nil
}

func (c *PaymentResultConsumer) Start(ctx context.Context) {
	msgs, err := c.channel.Consume(
		"payment_results",
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Failed to register consumer: %v", err)
		return
	}

	log.Println("PaymentResultConsumer started")

	for {
		select {
		case <-ctx.Done():
			log.Println("PaymentResultConsumer stopped")
			return
		case msg, ok := <-msgs:
			if !ok {
				return
			}
			c.processMessage(ctx, msg)
		}
	}
}

func (c *PaymentResultConsumer) processMessage(ctx context.Context, msg amqp.Delivery) {
	var result messaging.PaymentResultMessage
	if err := json.Unmarshal(msg.Body, &result); err != nil {
		log.Printf("Error unmarshaling message: %v", err)
		msg.Nack(false, true)
		return
	}

	cmd := commands.UpdateOrderStatusCommand{
		OrderID: result.OrderID,
		Success: result.Success,
	}

	if err := c.updateStatusHandler.Handle(ctx, cmd); err != nil {
		log.Printf("Error updating order status: %v", err)
		msg.Nack(false, true)
		return
	}

	msg.Ack(false)
	log.Printf("Payment result processed for order %s, success=%v", result.OrderID, result.Success)
}

func (c *PaymentResultConsumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
}
