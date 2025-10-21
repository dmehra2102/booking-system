package domain

import "time"

type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "pending"
	BookingStatusConfirmed BookingStatus = "confirmed"
	BookingStatusCancelled BookingStatus = "cancelled"
	BookingStatusCompleted BookingStatus = "completed"
	BookingStatusFailed    BookingStatus = "failed"
)

type Booking struct {
	ID            string        `json:"id" db:"id"`
	UserID        string        `json:"user_id" db:"user_id"`
	ResourceID    string        `json:"resource_id" db:"resource_id"`
	StartTime     time.Time     `json:"start_time" db:"start_time"`
	EndTime       time.Time     `json:"end_time" db:"end_time"`
	Status        BookingStatus `json:"status" db:"status"`
	Amount        float64       `json:"amount" db:"amount"`
	Currency      string        `json:"currency" db:"currency"`
	PaymentID     *string       `json:"payment_id,omitempty" db:"payment_id"`
	ReservationID *string       `json:"reservation_id,omitempty" db:"reservation_id"`
	Notes         string        `json:"notes,omitempty" db:"notes"`
	Metadata      string        `json:"metadata,omitempty" db:"metadata"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at" db:"updated_at"`
	UserName      string        `json:"user_name,omitempty" db:"user_name"`
	UserEmail     string        `json:"user_email,omitempty" db:"omitempty"`
	ResourceName  string        `json:"resource_name,omitempty" db:"resource_name"`
}

type CreateBookingRequest struct {
	UserID     string    `json:"user_id" validate:"required"`
	ResourceID string    `json:"resource_id" validate:"required"`
	StartTime  time.Time `json:"start_time" validate:"required"`
	EndTime    time.Time `json:"end_time" validate:"required"`
	Notes      string    `json:"notes,omitempty"`
}

type UpdateBookingRequest struct {
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Notes     *string    `json:"notes,omitempty"`
}

type CancelBookingRequest struct {
	Reason string `json:"reason" validate:"required"`
}

func (b *Booking) IsActive() bool {
	return b.Status == BookingStatusPending || b.Status == BookingStatusConfirmed
}

func (b *Booking) CanBeCancelled() bool {
	return b.Status == BookingStatusPending || b.Status == BookingStatusConfirmed
}

func (b *Booking) CanBeUpdated() bool {
	return b.Status == BookingStatusPending
}

func (b *Booking) Duration() time.Duration {
	return b.EndTime.Sub(b.StartTime)
}

func (b *Booking) IsOverlapping(other *Booking) bool {
	return b.ResourceID == other.ResourceID &&
		b.StartTime.Before(other.EndTime) &&
		b.EndTime.After(other.StartTime)
}
