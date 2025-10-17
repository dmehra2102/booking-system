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

type BookingRequestedEvent struct {
	BaseEvent
	Data BookingRequestedData `json:"data"`
}

type BookingRequestedData struct {
	BookingID  string    `json:"booking_id"`
	UserID     string    `json:"user_id"`
	ResourceID string    `json:"resource_id"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Amount     float64   `json:"amount"`
	Currency   string    `json:"currency"`
	Status     string    `json:"status"`
}

type BookingConfirmedEvent struct {
	BaseEvent
	Data BookingConfirmedData `json:"data"`
}

type BookingConfirmedData struct {
	BookingID   string    `json:"booking_id"`
	UserID      string    `json:"user_id"`
	ResourceID  string    `json:"resource_id"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	PaymentID   string    `json:"payment_id"`
	ConfirmedAt time.Time `json:"confirmed_at"`
}

type BookingCancelledEvent struct {
	BaseEvent
	Data BookingCancelledData `json:"data"`
}

type BookingCancelledData struct {
	BookingID   string    `json:"booking_id"`
	UserID      string    `json:"user_id"`
	ResourceID  string    `json:"resource_id"`
	Reason      string    `json:"reason"`
	CancelledAt time.Time `json:"cancelled_at"`
}

type InventoryReservedEvent struct {
	BaseEvent
	Data InventoryReservedData `json:"data"`
}

type InventoryReservedData struct {
	ResourceID    string    `json:"resource_id"`
	BookingID     string    `json:"booking_id"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	ReservedAt    time.Time `json:"reserved_at"`
	ReservationID string    `json:"reservation_id"`
}

type InventoryReleasedEvent struct {
	BaseEvent
	Data InventoryReleasedData `json:"data"`
}

type InventoryReleasedData struct {
	ResourceID    string    `json:"resource_id"`
	BookingID     string    `json:"booking_id"`
	ReservationID string    `json:"reservation_id"`
	ReleasedAt    time.Time `json:"released_at"`
	Reason        string    `json:"reason"`
}

// Payment Events
type PaymentProcessedEvent struct {
	BaseEvent
	Data PaymentProcessedData `json:"data"`
}

type PaymentProcessedData struct {
	PaymentID   string    `json:"payment_id"`
	BookingID   string    `json:"booking_id"`
	UserID      string    `json:"user_id"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	Method      string    `json:"method"`
	Status      string    `json:"status"`
	ProcessedAt time.Time `json:"processed_at"`
}

type PaymentFailedEvent struct {
	BaseEvent
	Data PaymentFailedData `json:"data"`
}

type PaymentFailedData struct {
	PaymentID string    `json:"payment_id"`
	BookingID string    `json:"booking_id"`
	UserID    string    `json:"user_id"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	Reason    string    `json:"reason"`
	FailedAt  time.Time `json:"failed_at"`
}

// Notification Events
type NotificationSentEvent struct {
	BaseEvent
	Data NotificationSentData `json:"data"`
}

type NotificationSentData struct {
	NotificationID string         `json:"notification_id"`
	UserID         string         `json:"user_id"`
	Type           string         `json:"type"`
	Channel        string         `json:"channel"`
	Subject        string         `json:"subject"`
	Content        string         `json:"content"`
	SentAt         time.Time      `json:"sent_at"`
	Metadata       map[string]any `json:"metadata,omitempty"`
}
