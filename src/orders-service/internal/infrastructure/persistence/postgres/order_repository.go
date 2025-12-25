package postgres

import (
	"context"
	"time"

	"github.com/kayotklimenko/gozon-go/orders-service/internal/domain/order"
	"gorm.io/gorm"
)

type OrderModel struct {
	ID          string    `gorm:"type:uuid;primaryKey"`
	UserID      string    `gorm:"type:uuid;index;not null"`
	Amount      float64   `gorm:"type:decimal(18,2);not null"`
	Description string    `gorm:"type:varchar(500);not null"`
	Status      string    `gorm:"type:varchar(20);not null;default:'New'"`
	CreatedAt   time.Time `gorm:"not null"`
	UpdatedAt   *time.Time
}

func (OrderModel) TableName() string {
	return "orders"
}

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Save(ctx context.Context, o *order.Order) error {
	db := GetDB(ctx, r.db)

	model := &OrderModel{
		ID:          string(o.ID()),
		UserID:      string(o.UserID()),
		Amount:      float64(o.Amount()),
		Description: o.Description(),
		Status:      string(o.Status()),
		CreatedAt:   o.CreatedAt(),
		UpdatedAt:   o.UpdatedAt(),
	}

	return db.Create(model).Error
}

func (r *OrderRepository) FindByID(ctx context.Context, id order.OrderID, userID order.UserID) (*order.Order, error) {
	db := GetDB(ctx, r.db)

	var model OrderModel
	err := db.Where("id = ? AND user_id = ?", string(id), string(userID)).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, order.ErrOrderNotFound
		}
		return nil, err
	}

	return r.toDomain(&model), nil
}

func (r *OrderRepository) FindByIDOnly(ctx context.Context, id order.OrderID) (*order.Order, error) {
	db := GetDB(ctx, r.db)

	var model OrderModel
	err := db.Where("id = ?", string(id)).First(&model).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, order.ErrOrderNotFound
		}
		return nil, err
	}

	return r.toDomain(&model), nil
}

func (r *OrderRepository) FindByUserID(ctx context.Context, userID order.UserID) ([]*order.Order, error) {
	db := GetDB(ctx, r.db)

	var models []OrderModel
	err := db.Where("user_id = ?", string(userID)).Order("created_at DESC").Find(&models).Error
	if err != nil {
		return nil, err
	}

	orders := make([]*order.Order, len(models))
	for i, model := range models {
		orders[i] = r.toDomain(&model)
	}

	return orders, nil
}

func (r *OrderRepository) Update(ctx context.Context, o *order.Order) error {
	db := GetDB(ctx, r.db)

	return db.Model(&OrderModel{}).
		Where("id = ? AND status = 'New'", string(o.ID())).
		Updates(map[string]interface{}{
			"status":     string(o.Status()),
			"updated_at": o.UpdatedAt(),
		}).Error
}

func (r *OrderRepository) toDomain(model *OrderModel) *order.Order {
	return order.ReconstructOrder(
		model.ID,
		model.UserID,
		model.Amount,
		model.Description,
		order.Status(model.Status),
		model.CreatedAt,
		model.UpdatedAt,
	)
}
