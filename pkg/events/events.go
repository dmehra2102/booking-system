package events

import (
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	UserCreated EventType = "user.created"
	UserUpdated EventType = "user.updated"
	UserDeleted EventType = "user.deleted"

	BookingRequested EventType = "booking.requested"
	BookingConfirmed EventType = "booking.confirmed"
	BookingCancelled EventType = "booking.cancelled"
	BookingUpdated   EventType = "booking.updated"

	InventoryReserved EventType = "inventory.reserved"
	InventoryReleased EventType = "inventory.released"
	InventoryUpdated  EventType = "inventory.updated"

	PaymentProcessed EventType = "payment.processed"
	PaymentFailed    EventType = "payment.failed"
	PaymentRefunded  EventType = "payment.refunded"

	NotificationSent   EventType = "notification.sent"
	NotificationFailed EventType = "notification.failed"
)

type BaseEvent struct {
	ID        string         `json:"id"`
	Type      EventType      `json:"type"`
	Source    string         `json:"source"`
	Timestamp time.Time      `json:"timestamp"`
	Version   string         `json:"version"`
	TraceID   string         `json:"trace_id,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

func NewBaseEvent(eventType EventType, source string, traceID string) BaseEvent {
	return BaseEvent{
		ID:        uuid.New().String(),
		Type:      eventType,
		Source:    source,
		Timestamp: time.Now().UTC(),
		Version:   "1.0",
		TraceID:   traceID,
		Metadata:  make(map[string]any),
	}
}

type UserCreatedEvent struct {
	BaseEvent
	Data UserCreatedData `json:"data"`
}

type UserCreatedData struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type UserUpdatedEvent struct {
	BaseEvent
	Data UserUpdatedData `json:"data"`
}

type UserUpdatedData struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserDeletedEvent struct {
	BaseEvent
	Data UserDeletedData `json:"data"`
}

type UserDeletedData struct {
	UserID    string    `json:"user_id"`
	DeletedAt time.Time `json:"deleted_at"`
}
