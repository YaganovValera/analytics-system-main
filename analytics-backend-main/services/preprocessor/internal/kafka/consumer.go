// preprocessor/internal/kafka/consumer.go

package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/YaganovValera/analytics-system/common/logger"
	marketdatapb "github.com/YaganovValera/analytics-system/proto/gen/go/v1/marketdata"
	"github.com/YaganovValera/analytics-system/services/preprocessor/internal/aggregator"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type ManualConsumer struct {
	topic      string
	groupID    string
	brokers    []string
	aggregator aggregator.Aggregator
	log        *logger.Logger
}

func NewManualConsumer(brokers []string, topic, groupID string, agg aggregator.Aggregator, log *logger.Logger) *ManualConsumer {
	return &ManualConsumer{
		topic:      topic,
		groupID:    groupID,
		brokers:    brokers,
		aggregator: agg,
		log:        log.Named("manual-consumer"),
	}
}

func (mc *ManualConsumer) Run(ctx context.Context) error {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Version = sarama.V2_8_0_0

	client, err := sarama.NewConsumerGroup(mc.brokers, mc.groupID, config)
	if err != nil {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}

	h := &consumerHandler{
		topic:      mc.topic,
		aggregator: mc.aggregator,
		log:        mc.log,
	}

	go func() {
		for err := range client.Errors() {
			mc.log.Error("consumer error", zap.Error(err))
		}
	}()

	for {
		if ctx.Err() != nil {
			mc.log.Info("context canceled, stopping consumer")
			return ctx.Err()
		}
		err := client.Consume(ctx, []string{mc.topic}, h)
		if err != nil {
			mc.log.Error("consume failed", zap.Error(err))
			time.Sleep(5 * time.Second)
		}
	}
}

type consumerHandler struct {
	topic      string
	aggregator aggregator.Aggregator
	log        *logger.Logger
}

func (h *consumerHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *consumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var data marketdatapb.MarketData
		if err := proto.Unmarshal(msg.Value, &data); err != nil {
			h.log.Error("failed to unmarshal MarketData", zap.ByteString("raw", msg.Value), zap.Error(err))
			continue
		}

		if err := h.aggregator.Process(session.Context(), &data); err != nil {
			h.log.Error("aggregator process failed", zap.Error(err))
		}
		session.MarkMessage(msg, "")
	}
	return nil
}
