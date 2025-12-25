package postgres

import (
	"context"
	"time"

	"github.com/kayotklimenko/gozon-go/payments-service/internal/domain/inbox"
	"gorm.io/gorm"
)

type InboxModel struct {
	ID             string    `gorm:"type:uuid;primaryKey"`
	IdempotencyKey string    `gorm:"type:varchar(100);uniqueIndex;not null"`
	MessageType    string    `gorm:"type:varchar(100);not null"`
	Payload        string    `gorm:"type:text"`
	Status         int       `gorm:"type:int;not null;default:0;index"`
	ReceivedAt     time.Time `gorm:"not null"`
	ProcessedAt    *time.Time
}

func (InboxModel) TableName() string {
	return "inbox_messages"
}

type InboxRepository struct {
	db *gorm.DB
}

func NewInboxRepository(db *gorm.DB) *InboxRepository {
	return &InboxRepository{db: db}
}

func (r *InboxRepository) Save(ctx context.Context, msg *inbox.Message) error {
	db := GetDB(ctx, r.db)

	model := &InboxModel{
		ID:             string(msg.ID()),
		IdempotencyKey: msg.IdempotencyKey(),
		MessageType:    msg.MessageType(),
		Payload:        msg.Payload(),
		Status:         int(msg.Status()),
		ReceivedAt:     msg.ReceivedAt(),
		ProcessedAt:    msg.ProcessedAt(),
	}

	return db.Create(model).Error
}

func (r *InboxRepository) FindByIdempotencyKey(ctx context.Context, key string) (*inbox.Message, error) {
	db := GetDB(ctx, r.db)

	var model InboxModel
	err := db.Where("idempotency_key = ?", key).First(&model).Error
	if err != nil {
		return nil, err
	}

	return inbox.ReconstructMessage(
		model.ID,
		model.IdempotencyKey,
		model.MessageType,
		model.Payload,
		inbox.Status(model.Status),
		model.ReceivedAt,
		model.ProcessedAt,
	), nil
}

func (r *InboxRepository) Update(ctx context.Context, msg *inbox.Message) error {
	db := GetDB(ctx, r.db)

	return db.Model(&InboxModel{}).
		Where("id = ?", string(msg.ID())).
		Updates(map[string]interface{}{
			"status":       int(msg.Status()),
			"processed_at": msg.ProcessedAt(),
		}).Error
}
