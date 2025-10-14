package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dmehra2102/booking-system/internal/common/logger"
	"github.com/dmehra2102/booking-system/internal/common/metrics"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/trace"
)

type MessageHandler func(ctx context.Context, key, value []byte, headers map[string]string) error

type Consumer struct {
	reader     *kafka.Reader
	logger     *logger.Logger
	metrics    *metrics.Metrics
	tracer     trace.Tracer
	handlers   map[string]MessageHandler
	maxRetries int
}

func NewConsumer(brokers []string, consumerGroup, topic string, logger *logger.Logger, metrics *metrics.Metrics, tracer trace.Tracer) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:          brokers,
		GroupID:          consumerGroup,
		Topic:            topic,
		MinBytes:         1,
		MaxBytes:         10e6,
		CommitInterval:   time.Second,
		StartOffset:      kafka.LastOffset,
		MaxWait:          1 * time.Second,
		ReadBatchTimeout: 10 * time.Second,
		Dialer: &kafka.Dialer{
			Timeout:   30 * time.Second,
			DualStack: true,
		},
		ErrorLogger: kafka.LoggerFunc(func(msg string, args ...any) {
			logger.Error(fmt.Sprintf("kafka consumer eroror: "+msg, args...))
		}),
	})

	return &Consumer{
		reader:     reader,
		logger:     logger,
		metrics:    metrics,
		tracer:     tracer,
		handlers:   make(map[string]MessageHandler),
		maxRetries: 3,
	}
}

func (c *Consumer) RegisterHandler(messageType string, handler MessageHandler) {
	c.handlers[messageType] = handler
}

func (c *Consumer) Start(ctx context.Context) error {
	c.logger.Info("starting kafka consumer")

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("consumer context cancelled, shutting down")
			return ctx.Err()
		default:
			err := c.processMessage(ctx)
			if err != nil {
				c.logger.WithError(err).Error("error processing message")
			}
		}
	}
}

func (c *Consumer) processMessage(ctx context.Context) error {
	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		c.metrics.MessageErrors.WithLabelValues(msg.Topic, "read").Inc()
		return fmt.Errorf("failed to read message: %w", err)
	}

	headers := make(map[string]string)
	for _, header := range msg.Headers {
		headers[string(header.Key)] = string(header.Value)
	}

	ctx, span := c.tracer.Start(ctx, fmt.Sprintf("kafka.consume.%s", msg.Topic))
	defer span.End()

	c.logger.WithContext(ctx).With("topic", msg.Topic).With("partition", fmt.Sprintf("%d", msg.Partition)).With("offset", fmt.Sprintf("%d", msg.Offset)).Debug("processing message")

	// Process message with retry logic
	err = c.processWithRetry(ctx, msg.Key, msg.Value, headers)
	if err != nil {
		c.metrics.MessageErrors.WithLabelValues(msg.Topic, "process").Inc()
		c.logger.WithContext(ctx).WithError(err).Error("failed to process message after retries")

		return err
	}

	c.metrics.MessagesConsumed.WithLabelValues(msg.Topic).Inc()
	return nil
}

func (c *Consumer) processWithRetry(ctx context.Context, key, value []byte, headers map[string]string) error {
	var err error

	for i := 0; i < c.maxRetries; i++ {
		messageType := headers["message-type"]
		if messageType == "" {
			var payload map[string]any
			if err := json.Unmarshal(value, &payload); err == nil {
				if mt, ok := payload["type"].(string); ok {
					messageType = mt
				}
			}
		}

		if handler, exists := c.handlers[messageType]; exists {
			err = handler(ctx, key, value, headers)
			if err == nil {
				return nil
			}
		} else {
			c.logger.WithContext(ctx).With("message_type", messageType).Warn("no handler found for message type")
			return fmt.Errorf("no handler found for message type: %s", messageType)
		}

		// Wait before retry with exponential backoff
		if i < c.maxRetries-1 {
			backoff := time.Duration(i+1) * time.Second
			c.logger.WithContext(ctx).With("attempt", fmt.Sprintf("%d", i+1)).With("backoff", backoff.String()).Warn("retrying message processing")

			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return fmt.Errorf("failed to process message after %d retries: %w", c.maxRetries, err)
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
