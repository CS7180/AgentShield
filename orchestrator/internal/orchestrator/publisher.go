package orchestrator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

const defaultScanStatusTopic = "agent.status"

// ScanEventPublisher publishes scan lifecycle/progress events.
type ScanEventPublisher interface {
	PublishScanStatus(ctx context.Context, scanID, status string, progress int, detail string) error
	Close() error
}

type NoopScanEventPublisher struct {
	logger *zap.Logger
}

func NewNoopScanEventPublisher(logger *zap.Logger) *NoopScanEventPublisher {
	return &NoopScanEventPublisher{logger: logger}
}

func (p *NoopScanEventPublisher) PublishScanStatus(_ context.Context, scanID, status string, progress int, detail string) error {
	p.logger.Debug("scan event publisher noop",
		zap.String("scan_id", scanID),
		zap.String("status", status),
		zap.Int("progress", progress),
		zap.String("detail", detail),
	)
	return nil
}

func (p *NoopScanEventPublisher) Close() error { return nil }

type KafkaScanEventPublisher struct {
	producer sarama.SyncProducer
	topic    string
	logger   *zap.Logger
}

func NewKafkaScanEventPublisherFromEnv(logger *zap.Logger) (*KafkaScanEventPublisher, error) {
	brokersRaw := strings.TrimSpace(os.Getenv("KAFKA_BROKERS"))
	if brokersRaw == "" {
		return nil, fmt.Errorf("KAFKA_BROKERS is required")
	}
	brokers := splitCSV(brokersRaw)
	if len(brokers) == 0 {
		return nil, fmt.Errorf("KAFKA_BROKERS has no valid values")
	}

	cfg := sarama.NewConfig()
	cfg.Version = sarama.V2_8_0_0
	cfg.Producer.Return.Successes = true
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Retry.Max = 3

	producer, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		return nil, fmt.Errorf("new kafka producer: %w", err)
	}

	topic := strings.TrimSpace(os.Getenv("ORCHESTRATOR_SCAN_STATUS_TOPIC"))
	if topic == "" {
		topic = defaultScanStatusTopic
	}

	return &KafkaScanEventPublisher{
		producer: producer,
		topic:    topic,
		logger:   logger,
	}, nil
}

func (p *KafkaScanEventPublisher) PublishScanStatus(_ context.Context, scanID, status string, progress int, detail string) error {
	msg := map[string]any{
		"scan_id": scanID,
		"data": map[string]any{
			"source":   "orchestrator",
			"status":   status,
			"progress": progress,
			"detail":   detail,
		},
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal scan status event: %w", err)
	}

	_, _, err = p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: p.topic,
		Value: sarama.ByteEncoder(payload),
	})
	if err != nil {
		return fmt.Errorf("publish kafka scan event: %w", err)
	}
	return nil
}

func (p *KafkaScanEventPublisher) Close() error {
	if p.producer == nil {
		return nil
	}
	return p.producer.Close()
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}
