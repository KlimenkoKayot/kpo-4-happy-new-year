package postgres

import (
	"context"

	"gorm.io/gorm"
)

type txKey struct{}

func GetDB(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return db
}
