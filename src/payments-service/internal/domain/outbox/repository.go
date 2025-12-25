package outbox

import "context"

type Repository interface {
	Save(ctx context.Context, message *Message) error
	FindPending(ctx context.Context, limit int) ([]*Message, error)
	Update(ctx context.Context, message *Message) error
}
