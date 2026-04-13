package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	pb "github.com/agentshield/api-gateway/proto/orchestrator"
	"github.com/agentshield/orchestrator/internal/orchestrator"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	logger, err := buildLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "fatal: build logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync() //nolint:errcheck

	port := os.Getenv("ORCHESTRATOR_PORT")
	if port == "" {
		port = "50051"
	}

	executor, err := orchestrator.NewExecutorFromEnv(logger)
	if err != nil {
		logger.Warn("failed to init real executor, falling back to noop", zap.Error(err))
	}
	var pipeline orchestrator.PipelineExecutor
	if executor != nil {
		pipeline = executor
		defer executor.Close()
	} else {
		pipeline = orchestrator.NewNoopExecutor(logger)
	}

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Fatal("listen", zap.Error(err), zap.String("port", port))
	}

	grpcServer := grpc.NewServer()
	pb.RegisterOrchestratorServiceServer(grpcServer, orchestrator.NewServer(pipeline, logger))

	go func() {
		logger.Info("orchestrator server listening", zap.String("port", port))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatal("serve grpc", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutdown signal received")
	grpcServer.GracefulStop()
	logger.Info("orchestrator stopped")
}

func buildLogger() (*zap.Logger, error) {
	env := os.Getenv("ENVIRONMENT")
	if env == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}
