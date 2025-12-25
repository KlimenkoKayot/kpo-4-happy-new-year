package inbox

import (
	"time"

	"github.com/google/uuid"
)

type MessageID string
type Status int

const (
	StatusReceived Status = iota
	StatusProcessed
	StatusFailed
)

type Message struct {
	id             MessageID
	idempotencyKey string
	messageType    string
	payload        string
	status         Status
	receivedAt     time.Time
	processedAt    *time.Time
}

func NewMessage(idempotencyKey, messageType, payload string) *Message {
	return &Message{
		id:             MessageID(uuid.New().String()),
		idempotencyKey: idempotencyKey,
		messageType:    messageType,
		payload:        payload,
		status:         StatusReceived,
		receivedAt:     time.Now().UTC(),
	}
}

func ReconstructMessage(id, idempotencyKey, messageType, payload string, status Status, receivedAt time.Time, processedAt *time.Time) *Message {
	return &Message{
		id:             MessageID(id),
		idempotencyKey: idempotencyKey,
		messageType:    messageType,
		payload:        payload,
		status:         status,
		receivedAt:     receivedAt,
		processedAt:    processedAt,
	}
}

func (m *Message) MarkAsProcessed() {
	m.status = StatusProcessed
	now := time.Now().UTC()
	m.processedAt = &now
}

func (m *Message) IsProcessed() bool {
	return m.status == StatusProcessed
}

func (m *Message) ID() MessageID           { return m.id }
func (m *Message) IdempotencyKey() string  { return m.idempotencyKey }
func (m *Message) MessageType() string     { return m.messageType }
func (m *Message) Payload() string         { return m.payload }
func (m *Message) Status() Status          { return m.status }
func (m *Message) ReceivedAt() time.Time   { return m.receivedAt }
func (m *Message) ProcessedAt() *time.Time { return m.processedAt }
