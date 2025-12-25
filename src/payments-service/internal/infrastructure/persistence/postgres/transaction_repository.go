package postgres

import (
	"context"
	"time"

	"github.com/kayotklimenko/gozon-go/payments-service/internal/domain/transaction"
	"gorm.io/gorm"
)

type TransactionModel struct {
	ID          string    `gorm:"type:uuid;primaryKey"`
	AccountID   string    `gorm:"type:uuid;index;not null"`
	OrderID     *string   `gorm:"type:uuid;uniqueIndex"`
	Amount      float64   `gorm:"type:decimal(18,2);not null"`
	Type        int       `gorm:"type:int;not null"`
	Description string    `gorm:"type:varchar(500)"`
	CreatedAt   time.Time `gorm:"not null"`
}

func (TransactionModel) TableName() string {
	return "transactions"
}

type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Save(ctx context.Context, tx *transaction.Transaction) error {
	db := GetDB(ctx, r.db)

	model := &TransactionModel{
		ID:          string(tx.ID()),
		AccountID:   tx.AccountID(),
		OrderID:     tx.OrderID(),
		Amount:      tx.Amount(),
		Type:        int(tx.Type()),
		Description: tx.Description(),
		CreatedAt:   tx.CreatedAt(),
	}

	return db.Create(model).Error
}

func (r *TransactionRepository) ExistsByOrderID(ctx context.Context, orderID string) (bool, error) {
	db := GetDB(ctx, r.db)

	var count int64
	err := db.Model(&TransactionModel{}).Where("order_id = ?", orderID).Count(&count).Error
	return count > 0, err
}
