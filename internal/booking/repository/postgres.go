package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dmehra2102/booking-system/internal/booking/domain"
	"github.com/dmehra2102/booking-system/internal/common/database"
	"github.com/dmehra2102/booking-system/internal/common/errors"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type PostgresBookingRepository struct {
	db     *database.PostgresDB
	tracer trace.Tracer
}

func NewPostgresBookingRepository(db *database.PostgresDB, tracer trace.Tracer) *PostgresBookingRepository {
	return &PostgresBookingRepository{
		db:     db,
		tracer: tracer,
	}
}

func (r *PostgresBookingRepository) Create(ctx context.Context, booking *domain.Booking) error {
	ctx, span := r.tracer.Start(ctx, "booking.repository.create")
	defer span.End()

	booking.ID = uuid.New().String()
	booking.CreatedAt = time.Now().UTC()
	booking.UpdatedAt = time.Now().UTC()

	query := `
		INSERT INTO bookings (
			id, user_id, resource_id, start_time, end_time, status,
			amount, currency, notes, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.db.Exec(ctx, query,
		booking.ID, booking.UserID, booking.ResourceID, booking.StartTime,
		booking.EndTime, booking.Status, booking.Amount, booking.Currency,
		booking.Notes, booking.Metadata, booking.CreatedAt, booking.UpdatedAt,
	)

	if err != nil {
		return errors.NewInternalError("failed to create booking", err)
	}

	return nil
}

func (r *PostgresBookingRepository) GetByID(ctx context.Context, id string) (*domain.Booking, error) {
	ctx,span := r.tracer.Start(ctx,"booking.repository.get_by_id")
	defer span.End()

	query := `
		SELECT b.id, b.user_id, b.resource_id, b.start_time, b.end_time, b.status,
				b.amount, b.currency, b.payment_id, b.reservation_id, b.notes,
				b.metadata, b.created_at, b.updated_at,
				u.name as user_name, u.email as user_email,
				r.name as resource_name
		FROM bookings b
		LEFT JOIN users u ON b.user_id = u.id
		LEFT JOIN resources r ON b.resource_id = r.id
		WHERE b.id = $1
	`

	booking := &domain.Booking{}
	var paymentID, reservationID sql.NullString
	var userName, userEmail, resourceName sql.NullString

	err := r.db.QueryRow(ctx, query, id).Scan(
		&booking.ID, &booking.UserID, &booking.ResourceID, &booking.StartTime,
		&booking.EndTime, &booking.Status, &booking.Amount, &booking.Currency,
		&paymentID, &reservationID, &booking.Notes, &booking.Metadata,
		&booking.CreatedAt, &booking.UpdatedAt,
		&userName, &userEmail, &resourceName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("booking")
		}
		return nil, errors.NewInternalError("failed to get boooking", err)
	}

	// Handle nullable fields
	if paymentID.Valid {
		booking.PaymentID = &paymentID.String
	}
	if reservationID.Valid {
		booking.ReservationID = &reservationID.String
	}
	if userName.Valid {
		booking.UserName = userName.String
	}
	if userEmail.Valid {
		booking.UserEmail = userEmail.String
	}
	if resourceName.Valid {
		booking.ResourceName = resourceName.String
	}

	return booking, nil
}

func (r *PostgresBookingRepository) Update(ctx context.Context, id string, updates map[string]any) error {
	ctx,span := r.tracer.Start(ctx,"booking.repository.update")
	defer span.End()

	if len(updates) == 0 {
		return nil
	}

	updates["updated_at"] = time.Now().UTC()

	setParts := make([]string, 0, len(updates))
	args := make([]any, 0, len(updates)+1)
	argIndex := 1

	for field, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = %d", field, argIndex))
		args = append(args, value)
		argIndex++
	}

	query := fmt.Sprintf("UPDATE bookings SET %s WHERE id = %d", joinStrings(setParts,", "), argIndex)
	args = append(args, id)

	result,err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return errors.NewInternalError("failed to update booking", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewInternalError("failed to check update result", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("booking")
	}

	return nil
}

func (r *PostgresBookingRepository) Delete(ctx context.Context, id string) error {
	ctx,span := r.tracer.Start(ctx, "booking.repository.delete")
	defer span.End()

	query :=  `DELETE FROM bookings WHERE id = $1`

	result,err := r.db.Exec(ctx,query,id)
	if err != nil {
		return errors.NewInternalError("failed to delete booking", err)
	}

	rowsAffected,err := result.RowsAffected()
	if err != nil {
		return errors.NewInternalError("failed to check delete result", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("booking")
	}

	return nil
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}