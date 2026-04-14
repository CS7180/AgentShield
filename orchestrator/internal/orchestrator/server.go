package orchestrator

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/agentshield/api-gateway/proto/orchestrator"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Server struct {
	pb.UnimplementedOrchestratorServiceServer
	manager  *Manager
	executor PipelineExecutor
	publisher ScanEventPublisher
	logger   *zap.Logger
}

func NewServer(executor PipelineExecutor, logger *zap.Logger) *Server {
	return NewServerWithPublisher(executor, NewNoopScanEventPublisher(logger), logger)
}

func NewServerWithPublisher(executor PipelineExecutor, publisher ScanEventPublisher, logger *zap.Logger) *Server {
	if executor == nil {
		executor = NewNoopExecutor(logger)
	}
	if publisher == nil {
		publisher = NewNoopScanEventPublisher(logger)
	}
	return &Server{
		manager:  NewManager(),
		executor: executor,
		publisher: publisher,
		logger:   logger,
	}
}

func (s *Server) StartScan(_ context.Context, req *pb.StartScanRequest) (*pb.StartScanResponse, error) {
	if req == nil {
		return &pb.StartScanResponse{Accepted: false, Message: "request is required"}, nil
	}
	if _, err := uuid.Parse(req.ScanId); err != nil {
		return &pb.StartScanResponse{Accepted: false, Message: "invalid scan_id"}, nil
	}
	if req.TargetEndpoint == "" {
		return &pb.StartScanResponse{Accepted: false, Message: "target_endpoint is required"}, nil
	}

	runCtx, accepted, message := s.manager.StartScan(req.ScanId)
	s.logger.Info("start scan", zap.String("scan_id", req.ScanId), zap.Bool("accepted", accepted), zap.String("mode", req.Mode), zap.Strings("attack_types", req.AttackTypes))
	if accepted {
		s.publishStatus(req.ScanId, ScanStatusRunning, 0, "scan accepted")
		execReq := ScanExecutionRequest{
			ScanID:         req.ScanId,
			TargetEndpoint: req.TargetEndpoint,
			Mode:           req.Mode,
			AttackTypes:    req.AttackTypes,
		}
		go s.runPipeline(runCtx, execReq)
	}
	return &pb.StartScanResponse{Accepted: accepted, Message: message}, nil
}

func (s *Server) StopScan(_ context.Context, req *pb.StopScanRequest) (*pb.StopScanResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}
	if _, err := uuid.Parse(req.ScanId); err != nil {
		return &pb.StopScanResponse{Stopped: false, Message: "invalid scan_id"}, nil
	}

	stopped, message := s.manager.StopScan(req.ScanId)
	if stopped {
		if err := s.executor.MarkStopped(context.Background(), req.ScanId); err != nil {
			s.logger.Warn("mark stopped in executor", zap.String("scan_id", req.ScanId), zap.Error(err))
		}
		_, progress := s.manager.GetStatus(req.ScanId)
		s.publishStatus(req.ScanId, ScanStatusStopped, progress, "scan stopped by request")
	}
	s.logger.Info("stop scan", zap.String("scan_id", req.ScanId), zap.Bool("stopped", stopped))
	return &pb.StopScanResponse{Stopped: stopped, Message: message}, nil
}

func (s *Server) ScanStatus(_ context.Context, req *pb.ScanStatusRequest) (*pb.ScanStatusResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}
	status, progress := s.manager.GetStatus(req.ScanId)
	return &pb.ScanStatusResponse{ScanId: req.ScanId, Status: status, Progress: int32(progress)}, nil
}

func (s *Server) runPipeline(ctx context.Context, req ScanExecutionRequest) {
	err := s.executor.Execute(ctx, req, func(progress int) {
		if setErr := s.manager.SetProgress(req.ScanID, progress); setErr != nil {
			s.logger.Warn("set pipeline progress", zap.String("scan_id", req.ScanID), zap.Error(setErr))
		}
		s.publishStatus(req.ScanID, ScanStatusRunning, progress, "pipeline progress update")
	})
	if err != nil {
		if errors.Is(err, context.Canceled) {
			s.logger.Info("scan pipeline canceled", zap.String("scan_id", req.ScanID))
			status, progress := s.manager.GetStatus(req.ScanID)
			s.publishStatus(req.ScanID, status, progress, "scan canceled")
			return
		}
		_ = s.manager.MarkFailed(req.ScanID)
		status, progress := s.manager.GetStatus(req.ScanID)
		s.publishStatus(req.ScanID, status, progress, err.Error())
		s.logger.Error("scan pipeline failed", zap.String("scan_id", req.ScanID), zap.Error(err))
		return
	}

	if err := s.manager.MarkCompleted(req.ScanID); err != nil {
		s.logger.Warn("mark completed in manager", zap.String("scan_id", req.ScanID), zap.Error(err))
	}
	s.publishStatus(req.ScanID, ScanStatusCompleted, 100, "scan pipeline completed")
	s.logger.Info("scan pipeline completed", zap.String("scan_id", req.ScanID))
}

func (s *Server) publishStatus(scanID, status string, progress int, detail string) {
	if s.publisher == nil {
		return
	}
	if err := s.publisher.PublishScanStatus(context.Background(), scanID, status, progress, detail); err != nil {
		s.logger.Warn("publish scan status event", zap.String("scan_id", scanID), zap.Error(err))
	}
}
