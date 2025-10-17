package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dmehra2102/booking-system/internal/common/database"
	"github.com/dmehra2102/booking-system/internal/common/errors"
	"github.com/dmehra2102/booking-system/internal/user/domain"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type PostgresUserRepository struct {
	db     *database.PostgresDB
	tracer trace.Tracer
}

func NewPostgresUserRepository(db *database.PostgresDB, tracer trace.Tracer) *PostgresUserRepository {
	return &PostgresUserRepository{db: db, tracer: tracer}
}

func (r *PostgresUserRepository) Create(ctx context.Context, user *domain.User) error {
	ctx, span := r.tracer.Start(ctx, "repository.create")
	defer span.End()

	user.ID = uuid.New().String()
	user.CreatedAt = time.Now().UTC()
	user.UpdatedAt = time.Now().UTC()
	user.Active = true
	user.Role = "user"

	query := `
		INSERT INTO users (id, email, name,password_hash, role, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Exec(ctx, query, user.ID, user.Email, user.Name, user.Password, user.Role, user.Active, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		if isDuplicateError(err) {
			return errors.NewConflictError("user with this email already exists")
		}
		return errors.NewInternalError("failed to create user", err)
	}

	return nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	ctx, span := r.tracer.Start(ctx, "user.repository.get_by_id")
	defer span.End()

	query := `
		SELECT id, email, name, password_hash, role, active, created_at, updated_at
		FROM users WHERE id = $1 AND active = true
	`

	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Name, &user.Password,
		&user.Role, &user.Active, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("user")
		}
		return nil, errors.NewInternalError("failed to get user", err)
	}

	return user, nil
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	ctx, span := r.tracer.Start(ctx, "user.repostiory.get_by_email")
	defer span.End()

	query := `
		SELECT id, email, name, password_hash, role, active, created_at, updated_at
		FROM users WHERE email = $1 AND active = true
	`

	user := &domain.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Name, &user.Password,
		&user.Role, &user.Active, &user.CreatedAt, &user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("user")
		}
		return nil, errors.NewInternalError("failed to get user", err)
	}

	return user, nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, id string, updates map[string]any) error {
	ctx, span := r.tracer.Start(ctx, "user.repository.update")
	defer span.End()

	if len(updates) == 0 {
		return nil
	}

	updates["updated_at"] = time.Now().UTC()

	setParts := make([]string, 0, len(updates))
	args := make([]any, 0, len(updates)+1)
	argIndex := 1

	for field, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
		args = append(args, value)
		argIndex++
	}

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d", joinStrings(setParts, ", "), argIndex)
	args = append(args, id)

	result, err := r.db.Exec(ctx, query, args...)
	if err != nil {
		return errors.NewInternalError("failed to update user", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewInternalError("failed to check update result", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("user")
	}

	return nil
}

func (r *PostgresUserRepository) Delete(ctx context.Context, id string) error {
	ctx, span := r.tracer.Start(ctx, "user.repository.delete")
	defer span.End()

	query := `UPDATE users SET active = false, updated_at = $1 WHERE id = $2`

	result, err := r.db.Exec(ctx, query, time.Now().UTC(), id)
	if err != nil {
		return errors.NewInternalError("failed to delete user", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.NewInternalError("failed to check delete result", err)
	}

	if rowsAffected == 0 {
		return errors.NewNotFoundError("user")
	}

	return nil
}

func (r *PostgresUserRepository) List(ctx context.Context, limit, offset int) ([]*domain.User, int64, error) {
	ctx, span := r.tracer.Start(ctx, "user.repository.list")
	defer span.End()

	countQuery := `SELECT COUNT(*) FROM users WHERE active = true`
	var total int64
	err := r.db.QueryRow(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, errors.NewInternalError("failed to count users", err)
	}

	query := `
		SELECT id, email, name, password_hash, role, active, created_at, updated_at
		FROM users WHERE active = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSER $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, errors.NewInternalError("failed to list users", err)
	}
	defer rows.Close()

	users := make([]*domain.User, 0)
	for rows.Next() {
		user := &domain.User{}
		err := rows.Scan(
			&user.ID, &user.Email, &user.Name, &user.Password,
			&user.Role, &user.Active, &user.CreatedAt, &user.UpdatedAt,
		)
		if err != nil {
			return nil, 0, errors.NewInternalError("failed to scan user", err)
		}
		users = append(users, user)
	}

	return users, total, nil
}

// Helper functions
func isDuplicateError(err error) bool {
	// PostgreSQL duplicate key error code is 23505
	return err != nil && (contains(err.Error(), "duplicate") ||
		contains(err.Error(), "unique constraint") ||
		contains(err.Error(), "23505"))
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
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
