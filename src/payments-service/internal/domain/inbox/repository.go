package inbox

import "context"

type Repository interface {
	Save(ctx context.Context, message *Message) error
	FindByIdempotencyKey(ctx context.Context, key string) (*Message, error)
	Update(ctx context.Context, message *Message) error
}
