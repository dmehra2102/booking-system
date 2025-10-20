package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dmehra2102/booking-system/internal/common/config"
	"github.com/dmehra2102/booking-system/internal/common/database"
	"github.com/dmehra2102/booking-system/internal/common/kafka"
	"github.com/dmehra2102/booking-system/internal/common/logger"
	"github.com/dmehra2102/booking-system/internal/common/metrics"
	"github.com/dmehra2102/booking-system/internal/common/middleware"
	"github.com/dmehra2102/booking-system/internal/common/tracing"
	"github.com/dmehra2102/booking-system/internal/user/handler"
	"github.com/dmehra2102/booking-system/internal/user/repository"
	"github.com/dmehra2102/booking-system/internal/user/service"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to load config: %v", err))
	}

	// Initialize logger
	log := logger.New(cfg.ServiceName, cfg.LogLevel)

	// Initialize tracing
	tracerShutdown := initTracing(cfg, log)
	defer tracerShutdown()

	tracer := noop.NewTracerProvider().Tracer(cfg.ServiceName)

	// Initialize metrics
	metricsCollector := metrics.New(cfg.ServiceName)

	// Initialize dependencies
	db := initDatabase(cfg, log, metricsCollector, tracer)
	defer db.Close()

	producer := kafka.NewProducer(cfg.KafkaBrokers, log, metricsCollector, tracer)
	defer producer.Close()

	// Initialize application components
	userRepo := repository.NewPostgresUserRepository(db, tracer)
	userService := service.NewUserService(
		userRepo,
		producer,
		log,
		metricsCollector,
		tracer,
		cfg.JWTSecret,
		cfg.JWTExpiry,
	)
	userHandler := handler.NewUserHandler(userService, log, tracer)

	// Setup router
	router := setupRouter(cfg, log, db, metricsCollector, userHandler)

	// Start server
	startServer(cfg, log, router)
}

// ------------------- Initialization Helpers -------------------

func initTracing(cfg *config.Config, log *logger.Logger) func() {
	tracerShutdown, err := tracing.InitTracer(cfg.ServiceName, cfg.JaegerEndpoint)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to initialize tracer: %v", err))
		return func() {}
	}
	return tracerShutdown
}

func initDatabase(cfg *config.Config, log *logger.Logger, m *metrics.Metrics, tracer trace.Tracer) *database.PostgresDB {
	db, err := database.NewPostgresDB(cfg.PostgresURL, log, m, tracer)
	if err != nil {
		log.Error(fmt.Sprintf("Failed to connect to database: %v", err))
		os.Exit(1)
	}
	return db
}

// ------------------- Router Setup -------------------

func setupRouter(cfg *config.Config, log *logger.Logger, db *database.PostgresDB, m *metrics.Metrics, userHandler *handler.UserHandler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Global middlewares
	router.Use(
		middleware.RequestID(),
		middleware.CORS(),
		middleware.Recovery(log),
		middleware.Timeout(30*time.Second),
		m.GinMiddleware(),
		otelgin.Middleware(cfg.ServiceName),
	)

	// Health Check
	router.GET("/health", func(ctx *gin.Context) {
		status := "healthy"
		dbStatus := "healthy"

		if err := db.Health(); err != nil {
			status = "unhealthy"
			dbStatus = "unhealthy"
		}

		ctx.JSON(http.StatusOK, gin.H{
			"status":   status,
			"database": dbStatus,
			"service":  cfg.ServiceName,
			"version":  "1.0.0",
		})
	})

	router.GET("/ready", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	// Metrics Endpoint
	router.GET("/metrics", gin.WrapH(m.Handler()))

	// API routes
	api := router.Group("/api/v1")
	{
		api.POST("/users", userHandler.CreateUser)
		api.POST("/auth/login", userHandler.Login)

		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			protected.GET("/users", userHandler.ListUsers)
			protected.GET("/users/:id", userHandler.GetUser)
			protected.PUT("/users/:id", userHandler.UpdateUser)
			protected.DELETE("/users/:id", userHandler.DeleteUser)
		}
	}

	return router
}

func startServer(cfg *config.Config, log *logger.Logger, router *gin.Engine) {
	server := &http.Server{
		Addr:    ":" + cfg.ServicePort,
		Handler: router,
	}

	go func() {
		log.Info(fmt.Sprintf("🚀 Starting %s on port %s", cfg.ServiceName, cfg.ServicePort))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error(fmt.Sprintf("Failed to start server: %v", err))
			os.Exit(1)
		}
	}()

	// Graceful shutdown
	waitForShutdown(server, log)
}

func waitForShutdown(server *http.Server, log *logger.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("🛑 Shutting down server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error(fmt.Sprintf("Server forced to shutdown: %v", err))
	}

	log.Info("✅ Server stopped cleanly")
}
