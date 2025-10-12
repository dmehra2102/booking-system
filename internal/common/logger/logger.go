package logger

import (
	"context"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/trace"
)

type Logger struct {
	logger zerolog.Logger
}

func New(serviceName, logLevel string) *Logger {

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	level := zerolog.InfoLevel
	switch strings.ToLower(logLevel) {
	case "debug":
		level = zerolog.DebugLevel
	case "info":
		level = zerolog.InfoLevel
	case "warn", "warning":
		level = zerolog.WarnLevel
	case "error":
		level = zerolog.ErrorLevel
	}

	logger := zerolog.New(os.Stdout).
		Level(level).
		With().
		Timestamp().Str("service", serviceName).
		Logger()

	return &Logger{logger: logger}
}

func (l *Logger) WithContext(ctx context.Context) *Logger {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		logger := l.logger.With().
			Str("trace_id", span.SpanContext().TraceID().String()).
			Str("span_id", span.SpanContext().SpanID().String()).
			Logger()
		return &Logger{logger: logger}
	}
	return l
}

func (l *Logger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

func (l *Logger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

func (l *Logger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

func (l *Logger) Error(msg string) {
	l.logger.Error().Msg(msg)
}

func (l *Logger) With(key, value string) *Logger {
	return &Logger{logger: l.logger.With().Str(key, value).Logger()}
}

func (l *Logger) WithError(err error) *Logger {
	return &Logger{logger: l.logger.With().Err(err).Logger()}
}

func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	event := l.logger.With()
	for k, v := range fields {
		event = event.Interface(k, v)
	}
	return &Logger{logger: event.Logger()}
}

var globalLogger *Logger

func Init(serviceName, logLevel string) {
	globalLogger = New(serviceName, logLevel)
}

func Debug(msg string) {
	if globalLogger != nil {
		globalLogger.Debug(msg)
	}
}

func Info(msg string) {
	if globalLogger != nil {
		globalLogger.Info(msg)
	}
}

func Warn(msg string) {
	if globalLogger != nil {
		globalLogger.Warn(msg)
	}
}

func Error(msg string) {
	if globalLogger != nil {
		globalLogger.Error(msg)
	}
}

func WithContext(ctx context.Context) *Logger {
	if globalLogger != nil {
		return globalLogger.WithContext(ctx)
	}
	return New("unknown", "info").WithContext(ctx)
}
