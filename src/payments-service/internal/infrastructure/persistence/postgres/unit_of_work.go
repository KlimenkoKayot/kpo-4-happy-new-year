package postgres

import (
	"context"

	"gorm.io/gorm"
)

type UnitOfWork struct {
	db *gorm.DB
}

func NewUnitOfWork(db *gorm.DB) *UnitOfWork {
	return &UnitOfWork{db: db}
}

func (u *UnitOfWork) Begin(ctx context.Context) (context.Context, error) {
	tx := u.db.Begin()
	if tx.Error != nil {
		return ctx, tx.Error
	}
	return context.WithValue(ctx, txKey{}, tx), nil
}

func (u *UnitOfWork) Commit(ctx context.Context) error {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.Commit().Error
	}
	return nil
}

func (u *UnitOfWork) Rollback(ctx context.Context) error {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.Rollback().Error
	}
	return nil
}
