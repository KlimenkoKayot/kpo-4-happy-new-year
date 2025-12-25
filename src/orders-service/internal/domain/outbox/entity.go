package outbox

import (
	"time"

	"github.com/google/uuid"
)

const MaxRetryCount = 5

type MessageID string
type MessageStatus int

const (
	StatusPending MessageStatus = iota
	StatusSent
	StatusFailed
)

type Message struct {
	id          MessageID
	messageType string
	payload     string
	status      MessageStatus
	createdAt   time.Time
	processedAt *time.Time
	retryCount  int
}

func NewMessage(messageType string, payload string) *Message {
	return &Message{
		id:          MessageID(uuid.New().String()),
		messageType: messageType,
		payload:     payload,
		status:      StatusPending,
		createdAt:   time.Now().UTC(),
		retryCount:  0,
	}
}

func ReconstructMessage(id, messageType, payload string, status MessageStatus, createdAt time.Time, processedAt *time.Time, retryCount int) *Message {
	return &Message{
		id:          MessageID(id),
		messageType: messageType,
		payload:     payload,
		status:      status,
		createdAt:   createdAt,
		processedAt: processedAt,
		retryCount:  retryCount,
	}
}

func (m *Message) MarkAsSent() {
	m.status = StatusSent
	now := time.Now().UTC()
	m.processedAt = &now
}

func (m *Message) IncrementRetry() {
	m.retryCount++
	if m.retryCount >= MaxRetryCount {
		m.status = StatusFailed
	}
}

func (m *Message) ID() MessageID           { return m.id }
func (m *Message) MessageType() string     { return m.messageType }
func (m *Message) Payload() string         { return m.payload }
func (m *Message) Status() MessageStatus   { return m.status }
func (m *Message) CreatedAt() time.Time    { return m.createdAt }
func (m *Message) ProcessedAt() *time.Time { return m.processedAt }
func (m *Message) RetryCount() int         { return m.retryCount }
func (m *Message) CanRetry() bool          { return m.retryCount < MaxRetryCount && m.status == StatusPending }
