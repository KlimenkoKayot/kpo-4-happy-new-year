package transaction

import "context"

type Repository interface {
	Save(ctx context.Context, tx *Transaction) error
	ExistsByOrderID(ctx context.Context, orderID string) (bool, error)
}
