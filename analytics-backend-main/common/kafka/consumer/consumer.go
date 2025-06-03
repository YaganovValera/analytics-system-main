// common/kafka/consumer/consumer.go
package consumer

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"

	"github.com/YaganovValera/analytics-system/common/backoff"
	commonkafka "github.com/YaganovValera/analytics-system/common/kafka"
	"github.com/YaganovValera/analytics-system/common/logger"
	"github.com/YaganovValera/analytics-system/common/serviceid"
)

func init() {
	serviceid.Register(SetServiceLabel)
}

// -----------------------------------------------------------------------------
// Service label (заполняется из common.InitServiceName)
// -----------------------------------------------------------------------------

var serviceLabel = "unknown"

// SetServiceLabel задаёт единое имя сервиса для метрик.
// Вызывается единожды из common.InitServiceName().
func SetServiceLabel(name string) { serviceLabel = name }

// -----------------------------------------------------------------------------
// Prometheus-метрики
// -----------------------------------------------------------------------------

var consumerMetrics = struct {
	ConnectAttempts *prometheus.CounterVec
	ConnectErrors   *prometheus.CounterVec
	ConsumeErrors   *prometheus.CounterVec
}{
	ConnectAttempts: promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "common", Subsystem: "kafka_consumer", Name: "connect_attempts_total",
			Help: "Kafka consumer group connect attempts",
		},
		[]string{"service"},
	),
	ConnectErrors: promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "common", Subsystem: "kafka_consumer", Name: "connect_errors_total",
			Help: "Kafka consumer connect errors",
		},
		[]string{"service"},
	),
	ConsumeErrors: promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "common", Subsystem: "kafka_consumer", Name: "consume_errors_total",
			Help: "Errors during consumption sessions",
		},
		[]string{"service"},
	),
}

// -----------------------------------------------------------------------------
// Consumer implementation
// -----------------------------------------------------------------------------

type kafkaConsumerGroup struct {
	group      sarama.ConsumerGroup
	log        *logger.Logger
	backoffCfg backoff.Config
}

// New создаёт и подключает ConsumerGroup с ретраями.
func New(ctx context.Context, cfg Config, log *logger.Logger) (commonkafka.Consumer, error) {
	log = log.Named("kafka-consumer")

	version, err := sarama.ParseKafkaVersion(cfg.Version)
	if err != nil {
		return nil, fmt.Errorf("kafka consumer: invalid Version %q: %w", cfg.Version, err)
	}
	sarCfg := sarama.NewConfig()
	sarCfg.Version = version
	sarCfg.Consumer.Return.Errors = true
	sarCfg.Consumer.Offsets.Initial = sarama.OffsetOldest

	var group sarama.ConsumerGroup
	connectOp := func(ctx context.Context) error {
		consumerMetrics.ConnectAttempts.WithLabelValues(serviceLabel).Inc()
		g, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.GroupID, sarCfg)
		if err != nil {
			consumerMetrics.ConnectErrors.WithLabelValues(serviceLabel).Inc()
			return err
		}
		group = g
		return nil
	}

	notify := func(ctx context.Context, err error, delay time.Duration, attempt int) {
		log.WithContext(ctx).Warn("kafka consumer retry",
			zap.Int("attempt", attempt),
			zap.Duration("delay", delay),
			zap.Error(err),
		)
	}

	if err := backoff.Execute(ctx, cfg.Backoff, connectOp, notify); err != nil {
		return nil, fmt.Errorf("kafka consumer: connect failed: %w", err)
	}

	log.Info("kafka consumer group connected",
		zap.Strings("brokers", cfg.Brokers),
		zap.String("group", cfg.GroupID),
	)
	return &kafkaConsumerGroup{group: group, log: log, backoffCfg: cfg.Backoff}, nil
}

// Consume запускает бесконечное чтение топиков, оборачивая сессии в backoff.
func (kc *kafkaConsumerGroup) Consume(
	ctx context.Context,
	topics []string,
	handler func(msg *commonkafka.Message) error,
) error {
	h := &consumerGroupHandler{handler: handler, log: kc.log}
	for {
		err := kc.group.Consume(ctx, topics, h)
		if err != nil {
			consumerMetrics.ConsumeErrors.WithLabelValues(serviceLabel).Inc()
			kc.log.Error("consume session error", zap.Error(err))

			// Небольшая пауза перед следующей сессией
			pause := func(ctx context.Context) error {
				select {
				case <-time.After(100 * time.Millisecond):
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			}

			notify := func(ctx context.Context, err error, delay time.Duration, attempt int) {
				kc.log.WithContext(ctx).Warn("pause between kafka sessions failed",
					zap.Int("attempt", attempt),
					zap.Duration("delay", delay),
					zap.Error(err),
				)
			}
			if berr := backoff.Execute(ctx, kc.backoffCfg, pause, notify); berr != nil {
				return fmt.Errorf("kafka consumer: pause between sessions failed: %w", berr)
			}
			continue
		}

		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

// Close закрывает ConsumerGroup.
func (kc *kafkaConsumerGroup) Close() error {
	return kc.group.Close()
}

// -----------------------------------------------------------------------------
// Internal handler
// -----------------------------------------------------------------------------

type consumerGroupHandler struct {
	handler func(msg *commonkafka.Message) error
	log     *logger.Logger
}

func (h *consumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case <-sess.Context().Done():
			return sess.Context().Err()
		case m, ok := <-claim.Messages():
			if !ok {
				return nil
			}

			ctxMsg := sess.Context()

			headers := make(map[string][]byte, len(m.Headers))
			for _, hdr := range m.Headers {
				if hdr != nil && hdr.Key != nil && hdr.Value != nil {
					headers[string(hdr.Key)] = hdr.Value
				}
			}

			msg := &commonkafka.Message{
				Key:       m.Key,
				Value:     m.Value,
				Topic:     m.Topic,
				Partition: m.Partition,
				Offset:    m.Offset,
				Timestamp: m.Timestamp,
				Headers:   headers,
			}

			// // ДО вызова h.handler(...)
			// h.log.WithContext(ctxMsg).Info("📩 Получено сообщение из Kafka",
			// 	zap.String("topic", m.Topic),
			// 	zap.String("key", string(m.Key)),
			// 	zap.Int32("partition", m.Partition),
			// 	zap.Int64("offset", m.Offset),
			// 	zap.Time("ts", m.Timestamp),
			// )

			if err := h.handler(msg); err != nil {
				h.log.WithContext(ctxMsg).Error("handler error", zap.Error(err))
			} else {
				sess.MarkMessage(m, "")
			}
		}
	}
}
