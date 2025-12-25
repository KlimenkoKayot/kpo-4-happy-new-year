package rabbitmq

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/kayotklimenko/gozon-go/payments-service/internal/application/commands"
	amqp "github.com/rabbitmq/amqp091-go"
)

type PaymentRequestMessage struct {
	OrderID        string    `json:"order_id"`
	UserID         string    `json:"user_id"`
	Amount         float64   `json:"amount"`
	CreatedAt      time.Time `json:"created_at"`
	IdempotencyKey string    `json:"idempotency_key"`
}

type PaymentRequestConsumer struct {
	channel        *amqp.Channel
	paymentHandler *commands.ProcessPaymentHandler
}

func NewPaymentRequestConsumer(conn *amqp.Connection, paymentHandler *commands.ProcessPaymentHandler) (*PaymentRequestConsumer, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &PaymentRequestConsumer{
		channel:        ch,
		paymentHandler: paymentHandler,
	}, nil
}

func (c *PaymentRequestConsumer) Start(ctx context.Context) {
	msgs, err := c.channel.Consume(
		"payment_requests",
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

	log.Println("PaymentRequestConsumer started")

	for {
		select {
		case <-ctx.Done():
			log.Println("PaymentRequestConsumer stopped")
			return
		case msg, ok := <-msgs:
			if !ok {
				return
			}
			c.processMessage(ctx, msg)
		}
	}
}

func (c *PaymentRequestConsumer) processMessage(ctx context.Context, msg amqp.Delivery) {
	var request PaymentRequestMessage
	if err := json.Unmarshal(msg.Body, &request); err != nil {
		log.Printf("Error unmarshaling message: %v", err)
		msg.Nack(false, true)
		return
	}

	cmd := commands.ProcessPaymentCommand{
		OrderID:        request.OrderID,
		UserID:         request.UserID,
		Amount:         request.Amount,
		IdempotencyKey: request.IdempotencyKey,
	}

	if err := c.paymentHandler.Handle(ctx, cmd); err != nil {
		log.Printf("Error processing payment: %v", err)
		msg.Nack(false, true)
		return
	}

	msg.Ack(false)
	log.Printf("Payment request processed for order %s", request.OrderID)
}

func (c *PaymentRequestConsumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
}
