package order

import "context"

type Repository interface {
	Save(ctx context.Context, order *Order) error
	FindByID(ctx context.Context, id OrderID, userID UserID) (*Order, error)
	FindByIDOnly(ctx context.Context, id OrderID) (*Order, error)
	FindByUserID(ctx context.Context, userID UserID) ([]*Order, error)
	Update(ctx context.Context, order *Order) error
}
