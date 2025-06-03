// preprocessor/internal/storage/kafkasink/kafkasink.go
package kafkasink

import (
	"context"
	"fmt"

	commonkafka "github.com/YaganovValera/analytics-system/common/kafka"
	"github.com/YaganovValera/analytics-system/common/logger"
	"github.com/YaganovValera/analytics-system/services/preprocessor/internal/aggregator"
	"github.com/YaganovValera/analytics-system/services/preprocessor/internal/metrics"
	"github.com/YaganovValera/analytics-system/services/preprocessor/internal/transport"

	"go.uber.org/zap"
)

// Sink публикует завершённые свечи в Kafka.
type Sink struct {
	producer commonkafka.Producer
	prefix   string
	log      *logger.Logger
}

// New создаёт новый Kafka sink.
func New(producer commonkafka.Producer, topicPrefix string, log *logger.Logger) *Sink {
	return &Sink{
		producer: producer,
		prefix:   topicPrefix,
		log:      log.Named("kafka-sink"),
	}
}

func (s *Sink) FlushCandle(ctx context.Context, c *aggregator.Candle) error {
	bytes, err := transport.MarshalCandleToKafka(ctx, c)
	if err != nil {
		s.log.WithContext(ctx).Error("marshal candle failed", zap.Error(err))
		metrics.KafkaPublishFailedTotal.WithLabelValues(c.Interval).Inc()
		return fmt.Errorf("kafka-sink: marshal: %w", err)
	}

	topic := fmt.Sprintf("%s.%s", s.prefix, c.Interval)
	key := []byte(c.Symbol)

	if err := s.producer.Publish(ctx, topic, key, bytes); err != nil {
		s.log.WithContext(ctx).Error("publish failed", zap.String("topic", topic), zap.Error(err))
		metrics.KafkaPublishFailedTotal.WithLabelValues(c.Interval).Inc()
		return fmt.Errorf("kafka-sink: publish: %w", err)
	}

	s.log.WithContext(ctx).Debug("published candle to Kafka",
		zap.String("symbol", c.Symbol),
		zap.String("interval", c.Interval),
		zap.String("topic", topic),
	)
	metrics.KafkaPublishedTotal.WithLabelValues(c.Interval).Inc()
	return nil
}
