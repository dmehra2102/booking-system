package service

import (
	"context"
	"time"

	"github.com/dmehra2102/booking-system/internal/common/kafka"
	"github.com/dmehra2102/booking-system/internal/common/logger"
	"github.com/dmehra2102/booking-system/internal/common/metrics"
	"github.com/dmehra2102/booking-system/internal/user/domain"
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
