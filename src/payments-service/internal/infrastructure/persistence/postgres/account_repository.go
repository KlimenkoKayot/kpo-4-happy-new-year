package postgres

import (
	"context"
	"time"

	"github.com/kayotklimenko/gozon-go/payments-service/internal/domain/account"
	"gorm.io/gorm"
)

type AccountModel struct {
	ID        string    `gorm:"type:uuid;primaryKey"`
	UserID    string    `gorm:"type:uuid;uniqueIndex;not null"`
	Balance   float64   `gorm:"type:decimal(18,2);not null;default:0"`
	Version   int       `gorm:"not null;default:1"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt *time.Time
}

func (AccountModel) TableName() string {
	return "accounts"
}

type AccountRepository struct {
	db *gorm.DB
}

func NewAccountRepository(db *gorm.DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) Save(ctx context.Context, acc *account.Account) error {
	db := GetDB(ctx, r.db)

	model := &AccountModel{
		ID:        string(acc.ID()),
		UserID:    string(acc.UserID()),
		Balance:   acc.Balance().Float64(),
		Version:   acc.Version(),
		CreatedAt: acc.CreatedAt(),
		UpdatedAt: acc.UpdatedAt(),
	}

	return db.Create(model).Error
}

func (r *AccountRepository) FindByUserID(ctx context.Context, userID account.UserID) (*account.Account, error) {
	db := GetDB(ctx, r.db)

	var model AccountModel
	err := db.Where("user_id = ?", string(userID)).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, account.ErrAccountNotFound
		}
		return nil, err
	}

	return r.toDomain(&model), nil
}

func (r *AccountRepository) Update(ctx context.Context, acc *account.Account) error {
	db := GetDB(ctx, r.db)

	result := db.Model(&AccountModel{}).
		Where("id = ? AND version = ?", string(acc.ID()), acc.Version()-1).
		Updates(map[string]interface{}{
			"balance":    acc.Balance().Float64(),
			"version":    acc.Version(),
			"updated_at": acc.UpdatedAt(),
		})

	if result.RowsAffected == 0 {
		return account.ErrConcurrencyConflict
	}

	return result.Error
}

func (r *AccountRepository) toDomain(model *AccountModel) *account.Account {
	return account.ReconstructAccount(
		model.ID,
		model.UserID,
		model.Balance,
		model.Version,
		model.CreatedAt,
		model.UpdatedAt,
	)
}
