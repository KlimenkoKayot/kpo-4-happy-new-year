package messaging

import "time"

type PaymentRequestMessage struct {
	OrderID        string    `json:"order_id"`
	UserID         string    `json:"user_id"`
	Amount         float64   `json:"amount"`
	CreatedAt      time.Time `json:"created_at"`
	IdempotencyKey string    `json:"idempotency_key"`
}

type PaymentResultMessage struct {
	OrderID      string    `json:"order_id"`
	Success      bool      `json:"success"`
	ErrorMessage string    `json:"error_message,omitempty"`
	ProcessedAt  time.Time `json:"processed_at"`
}
