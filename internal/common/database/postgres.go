package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/dmehra2102/booking-system/internal/common/logger"
	"github.com/dmehra2102/booking-system/internal/common/metrics"
	"go.opentelemetry.io/otel/trace"
)

type PostgresDB struct {
	db      *sql.DB
	logger  *logger.Logger
	metrics *metrics.Metrics
	tracer  trace.Tracer
}

func NewPostgresDB(url string, logger *logger.Logger, metrics *metrics.Metrics, tracer trace.Tracer) (*PostgresDB, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return &PostgresDB{
		db:      db,
		logger:  logger,
		metrics: metrics,
		tracer:  tracer,
	}, nil
}

func (p *PostgresDB) DB() *sql.DB {
	return p.db
}

func (p *PostgresDB) Close() error {
	return p.db.Close()
}

func (p *PostgresDB) Health() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return p.db.PingContext(ctx)
}

func (p *PostgresDB) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	ctx, span := p.tracer.Start(ctx, "postgres.query")
	defer span.End()

	start := time.Now()
	rows, err := p.db.QueryContext(ctx, query, args...)
	duration := time.Since(start).Seconds()

	if err != nil {
		p.metrics.DBQueries.WithLabelValues("query", "error").Inc()
		p.logger.WithContext(ctx).WithError(err).Error("database query failed")
		return nil, err
	}

	p.metrics.DBQueries.WithLabelValues("query", "success").Inc()
	p.metrics.DBQueryDuration.WithLabelValues("query").Observe(duration)

	return rows, nil
}

func (p *PostgresDB) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	ctx, span := p.tracer.Start(ctx, "postgres.query_row")
	defer span.End()

	start := time.Now()
	row := p.db.QueryRowContext(ctx, query, args...)
	duration := time.Since(start).Seconds()

	p.metrics.DBQueries.WithLabelValues("query", "success").Inc()
	p.metrics.DBQueryDuration.WithLabelValues("query").Observe(duration)

	return row
}

func (p *PostgresDB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	ctx, span := p.tracer.Start(ctx, "postgres.exec")
	defer span.End()

	start := time.Now()
	result, err := p.db.ExecContext(ctx, query, args...)
	duration := time.Since(start).Seconds()

	if err != nil {
		p.metrics.DBQueries.WithLabelValues("exec", "error").Inc()
		p.logger.WithContext(ctx).WithError(err).Error("database exec failed")
		return nil, err
	}

	p.metrics.DBQueries.WithLabelValues("exec", "success").Inc()
	p.metrics.DBQueryDuration.WithLabelValues("exec").Observe(duration)

	return result, nil
}

// BeginTx starts a new transaction
func (p *PostgresDB) BeginTx(ctx context.Context) (*sql.Tx, error) {
	ctx, span := p.tracer.Start(ctx, "postgres.begin_tx")
	defer span.End()

	return p.db.BeginTx(ctx, nil)
}
