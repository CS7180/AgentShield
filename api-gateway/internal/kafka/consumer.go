package kafka

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

var watchedTopics = []string{"agent.status", "judge.evaluations", "attack.results", "defense.results"}

// ConsumerGroup wraps a Sarama consumer group and routes messages to the Dispatcher.
type ConsumerGroup struct {
	client     sarama.ConsumerGroup
	dispatcher *Dispatcher
	logger     *zap.Logger
}

func NewConsumerGroup(brokers []string, groupID string, dispatcher *Dispatcher, logger *zap.Logger) (*ConsumerGroup, error) {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V2_8_0_0
	cfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}
	cfg.Consumer.Offsets.Initial = sarama.OffsetNewest

	client, err := sarama.NewConsumerGroup(brokers, groupID, cfg)
	if err != nil {
		return nil, fmt.Errorf("new consumer group: %w", err)
	}

	return &ConsumerGroup{
		client:     client,
		dispatcher: dispatcher,
		logger:     logger,
	}, nil
}

// Run starts the consumer loop. Blocks until ctx is cancelled.
func (cg *ConsumerGroup) Run(ctx context.Context) {
	handler := &consumerHandler{dispatcher: cg.dispatcher, logger: cg.logger}

	for {
		if err := cg.client.Consume(ctx, watchedTopics, handler); err != nil {
			if ctx.Err() != nil {
				return
			}
			cg.logger.Error("kafka consume error", zap.Error(err))
		}
		if ctx.Err() != nil {
			return
		}
	}
}

func (cg *ConsumerGroup) Close() error {
	return cg.client.Close()
}

// consumerHandler implements sarama.ConsumerGroupHandler.
type consumerHandler struct {
	dispatcher *Dispatcher
	logger     *zap.Logger
}

func (h *consumerHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *consumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			h.dispatcher.Dispatch(msg.Topic, msg.Value)
			session.MarkMessage(msg, "")

		case <-session.Context().Done():
			return nil
		}
	}
}
