package service

import (
	"context"
	"time"

	"github.com/dmehra2102/booking-system/internal/common/errors"
	"github.com/dmehra2102/booking-system/internal/common/kafka"
	"github.com/dmehra2102/booking-system/internal/common/logger"
	"github.com/dmehra2102/booking-system/internal/common/metrics"
	"github.com/dmehra2102/booking-system/internal/user/domain"
	"github.com/dmehra2102/booking-system/pkg/auth"
	"github.com/dmehra2102/booking-system/pkg/events"
	"github.com/dmehra2102/booking-system/pkg/validation"
	"go.opentelemetry.io/otel/trace"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, id string, updates map[string]any) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*domain.User, int64, error)
}

type UserService struct {
	repo      UserRepository
	producer  *kafka.Producer
	logger    *logger.Logger
	metrics   *metrics.Metrics
	tracer    trace.Tracer
	jwtSecret string
	jwtExpiry time.Duration
}

func NewUserService(
	repo UserRepository,
	producer *kafka.Producer,
	logger *logger.Logger,
	metrics *metrics.Metrics,
	tracer trace.Tracer,
	jwtSecret string,
	jwtExpiry time.Duration,
) *UserService {
	return &UserService{
		repo:      repo,
		producer:  producer,
		logger:    logger,
		metrics:   metrics,
		tracer:    tracer,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
}

func (s *UserService) CreateUser(ctx context.Context, req *domain.CreateUserRequest) (*domain.User, error) {
	ctx, span := s.tracer.Start(ctx, "user.service.create")
	defer span.End()

	// Validate Request
	if err := validation.ValidateStruct(req); err != nil {
		return nil, err
	}

	existingUser, err := s.repo.GetByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, errors.NewConflictError("user with this email already exists")
	}

	newUser := &domain.User{
		Email:    req.Email,
		Name:     req.Name,
		Password: req.Password,
	}

	if err := newUser.HashPassword(); err != nil {
		return nil, errors.NewInternalError("failed to hash password", err)
	}

	if err := s.repo.Create(ctx, newUser); err != nil {
		return nil, err
	}

	// Publish event
	event := events.UserCreatedEvent{
		BaseEvent: events.NewBaseEvent(events.UserCreated, "user-service", span.SpanContext().TraceID().String()),
		Data: events.UserCreatedData{
			UserID:    newUser.ID,
			Email:     newUser.Email,
			Name:      newUser.Name,
			CreatedAt: newUser.CreatedAt,
		},
	}

	if err := s.producer.Produce(ctx, string(events.UserCreated), newUser.ID, event); err != nil {
		s.logger.WithContext(ctx).WithError(err).Error("failed to publish user created event")
	}

	s.metrics.UsersTotal.WithLabelValues("created", "user").Inc()
	s.logger.WithContext(ctx).With("user_id", newUser.ID).Info("user created successfully")

	return newUser.ToPublic(), nil
}

func (s *UserService) Login(ctx context.Context, req *domain.LoginRequest) (*domain.LoginResponse, error) {
	ctx, span := s.tracer.Start(ctx, "user.service.login")
	defer span.End()

	// Validate Request
	if err := validation.ValidateStruct(req); err != nil {
		return nil, errors.NewValidationError("validation failed", err)
	}

	// Get user by email
	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.NewUnauthorizedError("invalid credentials")
	}

	// Check password
	if !user.CheckPassword(req.Password) {
		return nil, errors.NewUnauthorizedError("invalid credentials")
	}

	// Generate JWT token
	token, err := auth.GenerateToken(user.ID, user.Email, user.Role, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, errors.NewInternalError("failed to generate token", err)
	}

	response := &domain.LoginResponse{
		Token:     token,
		User:      user,
		ExpiresAt: time.Now().Add(s.jwtExpiry),
	}

	s.logger.WithContext(ctx).With("user_id", user.ID).Info("user logged in succcessfully")

	return response, nil
}

func (s *UserService) GetUser(ctx context.Context, id string) (*domain.User, error) {
	ctx, span := s.tracer.Start(ctx, "user.service.get")
	defer span.End()

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return user.ToPublic(), nil
}

func (s *UserService) UpdateUser(ctx context.Context, id string, req *domain.UpdateUserRequest) (*domain.User, error) {
	ctx, span := s.tracer.Start(ctx, "user.service.update")
	defer span.End()

	// validate request
	if err := validation.ValidateStruct(req); err != nil {
		return nil, errors.NewValidationError("validation failed", err)
	}

	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	updates := make(map[string]any)
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}

	if len(updates) == 0 {
		return s.GetUser(ctx, id)
	}

	if err := s.repo.Update(ctx, id, updates); err != nil {
		return nil, err
	}

	updatedUser, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	event := events.UserUpdatedEvent{
		BaseEvent: events.NewBaseEvent(events.UserUpdated, "user-service", span.SpanContext().TraceID().String()),
		Data: events.UserUpdatedData{
			UserID:    updatedUser.ID,
			Email:     updatedUser.Email,
			Name:      updatedUser.Name,
			UpdatedAt: updatedUser.UpdatedAt,
		},
	}

	if err := s.producer.Produce(ctx, string(events.UserUpdated), updatedUser.ID, event); err != nil {
		s.logger.WithContext(ctx).WithError(err).Error("failed to publish user updated event")
	}

	s.logger.WithContext(ctx).With("user_id", id).Info("user updated successfully")

	return updatedUser.ToPublic(), nil
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	ctx, span := s.tracer.Start(ctx, "user.service.delete")
	defer span.End()

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Publish event
	event := events.UserDeletedEvent{
		BaseEvent: events.NewBaseEvent(events.UserDeleted, "user-service", span.SpanContext().TraceID().String()),
		Data: events.UserDeletedData{
			UserID:    user.ID,
			DeletedAt: time.Now().UTC(),
		},
	}

	if err := s.producer.Produce(ctx, "user.deleted", user.ID, event); err != nil {
		s.logger.WithContext(ctx).WithError(err).Error("failed to publish user deleted event")
	}

	s.metrics.UsersDeleted.WithLabelValues("deleted", "user").Inc()
	s.logger.WithContext(ctx).With("user_id", id).Info("user deleted successfully")

	return nil
}

func (s *UserService) ListUsers(ctx context.Context, page, pageSize int) ([]*domain.User, int64, error) {
	ctx, span := s.tracer.Start(ctx, "user.service.list")
	defer span.End()

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	users, total, err := s.repo.List(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	publicUsers := make([]*domain.User, len(users))
	for i, user := range users {
		publicUsers[i] = user.ToPublic()
	}

	return publicUsers, total, nil
}
