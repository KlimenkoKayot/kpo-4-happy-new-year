package rabbitmq

import (
	"context"
	"log"
	"time"

	"github.com/kayotklimenko/gozon-go/payments-service/internal/domain/outbox"
	amqp "github.com/rabbitmq/amqp091-go"
)

type OutboxProcessor struct {
	repo    outbox.Repository
	channel *amqp.Channel
}

func NewOutboxProcessor(repo outbox.Repository, conn *amqp.Connection) (*OutboxProcessor, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &OutboxProcessor{
		repo:    repo,
		channel: ch,
	}, nil
}

func (p *OutboxProcessor) Start(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	log.Println("OutboxProcessor started")

	for {
		select {
		case <-ctx.Done():
			log.Println("OutboxProcessor stopped")
			return
		case <-ticker.C:
			p.processMessages(ctx)
		}
	}
}

func (p *OutboxProcessor) processMessages(ctx context.Context) {
	messages, err := p.repo.FindPending(ctx, 10)
	if err != nil {
		log.Printf("Error getting pending messages: %v", err)
		return
	}

	for _, msg := range messages {
		err := p.sendMessage(msg)
		if err != nil {
			log.Printf("Error sending message %s: %v", msg.ID(), err)
			msg.IncrementRetry()
			p.repo.Update(ctx, msg)
			continue
		}

		msg.MarkAsSent()
		if err := p.repo.Update(ctx, msg); err != nil {
			log.Printf("Error updating message status: %v", err)
		} else {
			log.Printf("Message %s sent successfully", msg.ID())
		}
	}
}

func (p *OutboxProcessor) sendMessage(msg *outbox.Message) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return p.channel.PublishWithContext(
		ctx,
		"payment_exchange",
		"payment_result",
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         []byte(msg.Payload()),
			MessageId:    string(msg.ID()),
		},
	)
}

func (p *OutboxProcessor) Close() {
	if p.channel != nil {
		p.channel.Close()
	}
}
