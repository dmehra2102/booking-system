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

type Producer struct {
	writer *kafka.Writer
	logger *logger.Logger
	metrics *metrics.Metrics
	tracer trace.Tracer
	maxRetries int
}

func NewProducer(brokers []string, logger *logger.Logger, metrics *metrics.Metrics, tracer trace.Tracer) *Producer {
	writer := &kafka.Writer{
		Addr: kafka.TCP(brokers...),
		Balancer: &kafka.LeastBytes{},
		BatchSize: 100,
		BatchTimeout: 10* time.Millisecond,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
		RequiredAcks: kafka.RequireAll,
		Async: false,
		Compression: kafka.Snappy,
		ErrorLogger: kafka.LoggerFunc(func(msg string, args ...any) {
			logger.Error(fmt.Sprintf("kafka producer error: " + msg, args))
		}),
	}

	return &Producer{
		writer: writer,
		logger: logger,
		metrics: metrics,
		tracer: tracer,
		maxRetries: 3,
	}
}

func (p *Producer) Produce(ctx context.Context, topic, key string, value any) error {
	ctx, span := p.tracer.Start(ctx, "kafka.produce")
	defer span.End()

	payload,err := json.Marshal(value)
	if err != nil {
		p.metrics.MessageErrors.WithLabelValues(topic, "serialization").Inc()
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	msg := kafka.Message{
		Topic: topic,
		Key: []byte(key),
		Value: payload,
		Time:  time.Now(),
		Headers: []kafka.Header{
			{Key: "content-type", Value: []byte("application/json")},
		},
	}

	if span.SpanContext().IsValid() {
		msg.Headers = append(msg.Headers, kafka.Header{
			Key: "trace-id",
			Value: []byte(span.SpanContext().TraceID().String()),
		})
	}

	err = p.writeWithRetry(ctx, msg)

	if err != nil {
		p.metrics.MessageErrors.WithLabelValues(topic, "produce").Inc()
		p.logger.WithContext(ctx).WithError(err).Error("failed to produce message")
		return fmt.Errorf("failed to produce message to topic %s: %w", topic, err)
	}

	p.metrics.MessagesProduced.WithLabelValues(topic).Inc()
	p.logger.WithContext(ctx).With("topic", topic).With("key", key).Debug("message produced successfully")

	return nil
}

func (p *Producer) writeWithRetry(ctx context.Context, msg kafka.Message) error {
	var err error
	for i:=0 ; i<p.maxRetries ; i++ {
		err := p.writer.WriteMessages(ctx,msg)
		if err == nil {
			return nil
		}

		if i<p.maxRetries - 1 {
			backoff := time.Duration(i+1) * time.Second
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return err
}

func (p *Producer) Close() error {
	return p.writer.Close()
}