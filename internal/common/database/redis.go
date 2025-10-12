package database

import (
	"context"
	"fmt"
	"time"

	"github.com/dmehra2102/booking-system/internal/common/logger"
	"github.com/dmehra2102/booking-system/internal/common/metrics"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/trace"
)

type RedisClient struct {
	client  *redis.Client
	logger  *logger.Logger
	metrics *metrics.Metrics
	tracer  trace.Tracer
}

func NewRedisClient(url string, logger *logger.Logger, metrics *metrics.Metrics, tracer trace.Tracer) (*RedisClient, error) {
	opt, err := redis.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping redis: %v", err)
	}

	return &RedisClient{
		client:  client,
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
	}, nil
}

func (r *RedisClient) Client() *redis.Client {
	return r.client
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}

func (r *RedisClient) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return r.client.Ping(ctx).Err()
}

func (r *RedisClient) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	ctx, span := r.tracer.Start(ctx, "redis.set")
	defer span.End()

	start := time.Now()
	err := r.client.Set(ctx, key, value, expiration).Err()
	duration := time.Since(start).Seconds()

	status := "success"
	if err != nil {
		status = "error"
		r.logger.WithContext(ctx).WithError(err).Error("redis set failed")
	}

	r.metrics.DBQueries.WithLabelValues("redis_set", status).Inc()
	r.metrics.DBQueryDuration.WithLabelValues("redis_set").Observe(duration)

	return err
}

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	ctx, span := r.tracer.Start(ctx, "redis.get")
	defer span.End()

	start := time.Now()
	result, err := r.client.Get(ctx, key).Result()
	duration := time.Since(start).Seconds()

	status := "success"
	if err != nil && err != redis.Nil {
		status = "error"
		r.logger.WithContext(ctx).WithError(err).Error("redis got failed")
	}

	r.metrics.DBQueries.WithLabelValues("redis_get", status).Inc()
	r.metrics.DBQueryDuration.WithLabelValues("redis_get").Observe(duration)

	return result, err
}

func (r *RedisClient) Delete(ctx context.Context, keys ...string) error {
	ctx, span := r.tracer.Start(ctx, "redis.delete")
	defer span.End()

	start := time.Now()
	err := r.client.Del(ctx, keys...).Err()
	duration := time.Since(start).Seconds()

	status := "success"
	if err != nil {
		status = "error"
		r.logger.WithContext(ctx).WithError(err).Error("redis delete failed")
	}

	r.metrics.DBQueries.WithLabelValues("redis_delete", status).Inc()
	r.metrics.DBQueryDuration.WithLabelValues("redis_delete").Observe(duration)

	return err
}
