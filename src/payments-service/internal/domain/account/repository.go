package account

import "context"

type Repository interface {
	Save(ctx context.Context, account *Account) error
	FindByUserID(ctx context.Context, userID UserID) (*Account, error)
	Update(ctx context.Context, account *Account) error
}
