package orchestrator

import (
	"context"
	"fmt"
	"time"

	pb "github.com/agentshield/api-gateway/proto/orchestrator"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// Client is the interface the handlers use to call the orchestrator.
type Client interface {
	StartScan(ctx context.Context, scanID, targetEndpoint, mode string, attackTypes []string) (accepted bool, message string, err error)
	StopScan(ctx context.Context, scanID string) (stopped bool, message string, err error)
}

// GRPCClient implements Client by dialing the real orchestrator over gRPC.
type GRPCClient struct {
	conn   *grpc.ClientConn
	stub   pb.OrchestratorServiceClient
	logger *zap.Logger
}

// NewGRPCClient creates a lazy-dial gRPC connection to addr.
func NewGRPCClient(addr string, logger *zap.Logger) (*GRPCClient, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("grpc dial %s: %w", addr, err)
	}

	return &GRPCClient{
		conn:   conn,
		stub:   pb.NewOrchestratorServiceClient(conn),
		logger: logger,
	}, nil
}

func (c *GRPCClient) StartScan(ctx context.Context, scanID, targetEndpoint, mode string, attackTypes []string) (bool, string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	resp, err := c.stub.StartScan(ctx, &pb.StartScanRequest{
		ScanId:         scanID,
		TargetEndpoint: targetEndpoint,
		Mode:           mode,
		AttackTypes:    attackTypes,
	})
	if err != nil {
		if isUnavailable(err) {
			c.logger.Warn("orchestrator unavailable, queueing scan", zap.String("scan_id", scanID))
			return false, "orchestrator unavailable", nil
		}
		return false, "", fmt.Errorf("StartScan rpc: %w", err)
	}
	return resp.Accepted, resp.Message, nil
}

func (c *GRPCClient) StopScan(ctx context.Context, scanID string) (bool, string, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	resp, err := c.stub.StopScan(ctx, &pb.StopScanRequest{ScanId: scanID})
	if err != nil {
		if isUnavailable(err) {
			return false, "orchestrator unavailable", nil
		}
		return false, "", fmt.Errorf("StopScan rpc: %w", err)
	}
	return resp.Stopped, resp.Message, nil
}

func (c *GRPCClient) Close() error {
	return c.conn.Close()
}

func isUnavailable(err error) bool {
	st, ok := status.FromError(err)
	return ok && (st.Code() == codes.Unavailable || st.Code() == codes.DeadlineExceeded)
}
