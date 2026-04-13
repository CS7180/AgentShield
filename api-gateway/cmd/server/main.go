package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/agentshield/api-gateway/internal/config"
	"github.com/agentshield/api-gateway/internal/handler"
	"github.com/agentshield/api-gateway/internal/kafka"
	"github.com/agentshield/api-gateway/internal/middleware"
	"github.com/agentshield/api-gateway/internal/orchestrator"
	pgrepo "github.com/agentshield/api-gateway/internal/repository/postgres"
	redisrepo "github.com/agentshield/api-gateway/internal/repository/redis"
	"github.com/agentshield/api-gateway/internal/storage"
	"github.com/agentshield/api-gateway/internal/ws"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const version = "0.1.0"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// ── Logger ──────────────────────────────────────────────────────────────────
	logger, err := buildLogger()
	if err != nil {
		return fmt.Errorf("build logger: %w", err)
	}
	defer logger.Sync() //nolint:errcheck

	// ── Config ──────────────────────────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// ── PostgreSQL ──────────────────────────────────────────────────────────────
	ctx := context.Background()
	pool, err := pgrepo.NewPool(ctx, cfg.Database.URL, cfg.Database.MaxConns)
	if err != nil {
		return fmt.Errorf("postgres pool: %w", err)
	}
	defer pool.Close()
	logger.Info("postgres connected")

	// Run migrations
	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}
	if err := pgrepo.RunMigrations(ctx, pool, migrationsDir, logger); err != nil {
		logger.Warn("migrations failed (may need manual SQL execution)", zap.Error(err))
	}

	// ── Redis ───────────────────────────────────────────────────────────────────
	redisClient, err := redisrepo.NewClient(cfg.Redis.URL)
	if err != nil {
		return fmt.Errorf("redis client: %w", err)
	}
	defer redisClient.Close()
	logger.Info("redis connected")

	// ── Repositories ────────────────────────────────────────────────────────────
	scanRepo := pgrepo.NewScanRepository(pool)
	reportRepo := pgrepo.NewReportRepository(pool)
	attackResultRepo := pgrepo.NewAttackResultRepository(pool)
	judgeCalibrationRepo := pgrepo.NewJudgeCalibrationRepository(pool)
	ownershipAdapter := pgrepo.NewOwnershipAdapter(scanRepo)

	// ── Storage ─────────────────────────────────────────────────────────────────
	var reportUploader storage.Uploader
	if cfg.Supabase.URL != "" && cfg.Supabase.ServiceRoleKey != "" {
		reportUploader = storage.NewSupabaseUploader(cfg.Supabase.URL, cfg.Supabase.ServiceRoleKey)
	} else {
		logger.Warn("supabase storage uploader disabled; report upsert endpoint will return NOT_IMPLEMENTED")
	}

	// ── Orchestrator ─────────────────────────────────────────────────────────────
	var orchClient handler.OrchestratorClient
	if cfg.Orchestrator.Enabled {
		grpcClient, err := orchestrator.NewGRPCClient(cfg.Orchestrator.Addr, logger)
		if err != nil {
			logger.Warn("orchestrator grpc init failed, using stub", zap.Error(err))
			orchClient = orchestrator.NewStub()
		} else {
			orchClient = grpcClient
			defer grpcClient.Close()
			logger.Info("orchestrator gRPC client ready", zap.String("addr", cfg.Orchestrator.Addr))
		}
	} else {
		orchClient = orchestrator.NewStub()
		logger.Info("orchestrator disabled, using in-process stub")
	}

	// ── WebSocket Hub ────────────────────────────────────────────────────────────
	hub := ws.NewHub(logger)
	go hub.Run()

	// ── Kafka Consumer ───────────────────────────────────────────────────────────
	dispatcher := kafka.NewDispatcher(hub, logger)
	consumerGroup, err := kafka.NewConsumerGroup(cfg.Kafka.Brokers, cfg.Kafka.GroupID, dispatcher, logger)
	if err != nil {
		logger.Warn("kafka consumer group init failed, WebSocket feed disabled", zap.Error(err))
	} else {
		kafkaCtx, kafkaCancel := context.WithCancel(ctx)
		defer kafkaCancel()
		go consumerGroup.Run(kafkaCtx)
		defer consumerGroup.Close()
		logger.Info("kafka consumer group started", zap.Strings("brokers", cfg.Kafka.Brokers))
	}

	// ── Gin Router ───────────────────────────────────────────────────────────────
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	jwtSecret := []byte(cfg.Supabase.JWTSecret)

	// Global middleware
	r.Use(
		middleware.RequestID(),
		middleware.Recovery(logger),
		middleware.CORS(cfg.CORS.AllowedOrigins),
		middleware.Logger(logger),
	)

	// Health + metrics (no auth)
	healthHandler := handler.NewHealthHandler(version)
	r.GET("/health", healthHandler.Health)
	r.GET("/metrics", healthHandler.Metrics())

	// WebSocket (auth via ?token= query param — jwt_auth middleware skipped here)
	wsHandler := handler.NewWSHandler(hub, jwtSecret, logger)
	r.GET("/ws/scans/:id/status", wsHandler.HandleScanStatus)

	// Authenticated API routes
	scanHandler := handler.NewScanHandler(scanRepo, orchClient, logger)
	attackResultHandler := handler.NewAttackResultHandler(attackResultRepo, logger)
	reportHandler := handler.NewReportHandler(reportRepo, reportUploader, cfg.Supabase.ReportsBucket, logger)
	reportGenerationHandler := handler.NewReportGenerationHandler(
		reportRepo,
		attackResultRepo,
		scanRepo,
		reportUploader,
		cfg.Supabase.ReportsBucket,
		logger,
	)
	judgeHandler := handler.NewJudgeHandler(judgeCalibrationRepo)

	globalRateLimit := middleware.GlobalRateLimit(redisClient)
	scanCreateRateLimit := middleware.ScanCreateRateLimit(redisClient)
	ownership := middleware.Ownership(ownershipAdapter)

	api := r.Group("/api/v1")
	api.Use(middleware.JWTAuth(jwtSecret), globalRateLimit)
	{
		scans := api.Group("/scans")
		{
			scans.POST("", scanCreateRateLimit, scanHandler.Create)
			scans.GET("", scanHandler.List)
			scans.GET("/:id", ownership, scanHandler.Get)
			scans.POST("/:id/start", ownership, scanHandler.Start)
			scans.POST("/:id/stop", ownership, scanHandler.Stop)
			scans.POST("/:id/attack-results", ownership, attackResultHandler.CreateBatch)
			scans.GET("/:id/attack-results", ownership, attackResultHandler.List)
			scans.PUT("/:id/report", ownership, reportHandler.Upsert)
			scans.GET("/:id/report", ownership, reportHandler.GetJSON)
			scans.GET("/:id/report/pdf", ownership, reportHandler.GetPDF)
			scans.POST("/:id/report/generate", ownership, reportGenerationHandler.Generate)
			scans.GET("/:id/compare/:other_id", ownership, reportHandler.Compare)
		}

		judge := api.Group("/judge")
		{
			judge.POST("/calibrate", judgeHandler.Calibrate)
			judge.GET("/calibration-report", judgeHandler.CalibrationReport)
		}
	}

	// ── HTTP Server ──────────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("server starting", zap.String("port", cfg.Server.Port), zap.String("env", cfg.Server.Environment))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	// ── Graceful Shutdown ────────────────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		return fmt.Errorf("server error: %w", err)
	case sig := <-quit:
		logger.Info("shutdown signal received", zap.String("signal", sig.String()))
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	logger.Info("server stopped gracefully")
	return nil
}

func buildLogger() (*zap.Logger, error) {
	env := os.Getenv("ENVIRONMENT")
	if env == "production" {
		return zap.NewProduction()
	}

	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	return cfg.Build()
}
