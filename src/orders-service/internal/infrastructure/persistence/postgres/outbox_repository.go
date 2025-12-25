package postgres

import (
	"context"
	"time"

	"github.com/kayotklimenko/gozon-go/orders-service/internal/domain/outbox"
	"gorm.io/gorm"
)

type OutboxModel struct {
	ID          string    `gorm:"type:uuid;primaryKey"`
	MessageType string    `gorm:"type:varchar(100);not null"`
	Payload     string    `gorm:"type:text;not null"`
	Status      int       `gorm:"type:int;not null;default:0;index"`
	CreatedAt   time.Time `gorm:"not null;index"`
	ProcessedAt *time.Time
	RetryCount  int `gorm:"not null;default:0"`
}

func (OutboxModel) TableName() string {
	return "outbox_messages"
}

type OutboxRepository struct {
	db *gorm.DB
}

func NewOutboxRepository(db *gorm.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

func (r *OutboxRepository) Save(ctx context.Context, msg *outbox.Message) error {
	db := GetDB(ctx, r.db)

	model := &OutboxModel{
		ID:          string(msg.ID()),
		MessageType: msg.MessageType(),
		Payload:     msg.Payload(),
		Status:      int(msg.Status()),
		CreatedAt:   msg.CreatedAt(),
		ProcessedAt: msg.ProcessedAt(),
		RetryCount:  msg.RetryCount(),
	}

	return db.Create(model).Error
}

func (r *OutboxRepository) FindPending(ctx context.Context, limit int) ([]*outbox.Message, error) {
	var models []OutboxModel

	err := r.db.Raw(`
		SELECT * FROM outbox_messages
		WHERE status = ? AND retry_count < ?
		ORDER BY created_at
		LIMIT ?
		FOR UPDATE SKIP LOCKED
	`, int(outbox.StatusPending), outbox.MaxRetryCount, limit).Scan(&models).Error

	if err != nil {
		return nil, err
	}

	messages := make([]*outbox.Message, len(models))
	for i, model := range models {
		messages[i] = outbox.ReconstructMessage(
			model.ID,
			model.MessageType,
			model.Payload,
			outbox.MessageStatus(model.Status),
			model.CreatedAt,
			model.ProcessedAt,
			model.RetryCount,
		)
	}

	return messages, nil
}

func (r *OutboxRepository) Update(ctx context.Context, msg *outbox.Message) error {
	return r.db.Model(&OutboxModel{}).
		Where("id = ?", string(msg.ID())).
		Updates(map[string]interface{}{
			"status":       int(msg.Status()),
			"processed_at": msg.ProcessedAt(),
			"retry_count":  msg.RetryCount(),
		}).Error
}
