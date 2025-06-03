package app

import (
	"context"
	"fmt"
	"time"

	"github.com/YaganovValera/analytics-system/common/httpserver"
	kafkamsg "github.com/YaganovValera/analytics-system/common/kafka"
	"github.com/YaganovValera/analytics-system/common/kafka/consumer"
	"github.com/YaganovValera/analytics-system/common/kafka/producer"
	"github.com/YaganovValera/analytics-system/common/logger"
	commonredis "github.com/YaganovValera/analytics-system/common/redis"
	"github.com/YaganovValera/analytics-system/common/serviceid"

	"github.com/YaganovValera/analytics-system/services/preprocessor/internal/aggregator"
	"github.com/YaganovValera/analytics-system/services/preprocessor/internal/config"
	"github.com/YaganovValera/analytics-system/services/preprocessor/internal/kafka"
	"github.com/YaganovValera/analytics-system/services/preprocessor/internal/metrics"
	kafkasink "github.com/YaganovValera/analytics-system/services/preprocessor/internal/storage/kafkasink"
	redisadapter "github.com/YaganovValera/analytics-system/services/preprocessor/internal/storage/redis"
	"github.com/YaganovValera/analytics-system/services/preprocessor/internal/storage/timescaledb"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func Run(ctx context.Context, cfg *config.Config, log *logger.Logger) error {
	serviceid.InitServiceName(cfg.ServiceName)
	metrics.Register(nil)

	// === Инициализация Redis ===
	rcli, err := commonredis.New(cfg.Redis, log)
	if err != nil {
		return fmt.Errorf("redis init: %w", err)
	}

	defer shutdownSafe(ctx, "redis", func(ctx context.Context) error {
		return rcli.Close()
	}, log)

	storage := redisadapter.NewAdapter(rcli)

	// === Применяем миграции TimescaleDB ===
	log.WithContext(ctx).Info("🔌 [Init] Применяем миграции TimescaleDB...")
	if err := timescaledb.ApplyMigrations(cfg.Timescale, log); err != nil {
		return fmt.Errorf("timescaledb migrations: %w", err)
	}

	tsWriter, err := timescaledb.NewTimescaleWriter(cfg.Timescale, log)
	if err != nil {
		return fmt.Errorf("timescaledb init: %w", err)
	}

	defer shutdownSafe(ctx, "timescaledb", func(ctx context.Context) error {
		tsWriter.Close()
		return nil
	}, log)

	// === Kafka Producer ===
	kprod, err := producer.New(ctx, cfg.KafkaProducer, log)
	if err != nil {
		return fmt.Errorf("kafka producer init: %w", err)
	}

	defer shutdownSafe(ctx, "kafka-producer", func(ctx context.Context) error {
		return kprod.Close()
	}, log)

	// === Aggregator ===
	kSink := kafkasink.New(kprod, cfg.OutputTopicPrefix, log)
	flushSink := aggregator.NewMultiSink(tsWriter, kSink)
	agg, err := aggregator.NewManager(cfg.Intervals, flushSink, storage, log)
	if err != nil {
		return fmt.Errorf("aggregator init: %w", err)
	}

	defer shutdownSafe(ctx, "aggregator", func(ctx context.Context) error {
		return agg.Close()
	}, log)

	// === Kafka Consumer ===
	kcons, err := consumer.New(ctx, cfg.KafkaConsumer, log)
	if err != nil {
		log.WithContext(ctx).Error(" Kafka Consumer init failed", zap.Error(err))
		return fmt.Errorf("kafka consumer init: %w", err)
	}
	log.WithContext(ctx).Info(" Kafka Consumer готов")

	defer shutdownSafe(ctx, "kafka-consumer", func(ctx context.Context) error {
		return kcons.Close()
	}, log)

	// === HTTP Server ===
	log.WithContext(ctx).Info("🌐 [Init] Запуск HTTP-сервера метрик...")
	readiness := func() error {
		ctxPing, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		return tsWriter.Ping(ctxPing)
	}
	httpSrv, err := httpserver.New(
		cfg.HTTP,
		readiness,
		log,
		nil,
		httpserver.RecoverMiddleware,
		httpserver.CORSMiddleware(),
	)
	if err != nil {
		log.WithContext(ctx).Error(" HTTP-сервер init failed", zap.Error(err))
		return fmt.Errorf("httpserver init: %w", err)
	}
	log.WithContext(ctx).Info("HTTP-сервер готов")

	// === Группа запуска goroutines ===
	g, ctx := errgroup.WithContext(ctx)

	// === HTTP Server ===
	g.Go(func() error {
		log.WithContext(ctx).Info("Запускаем HTTP-сервер", zap.Int("port", cfg.HTTP.Port))
		return httpSrv.Run(ctx)
	})

	// === Кастомный Consumer свечей (MarketData) ===
	manualConsumer := kafka.NewManualConsumer(
		cfg.KafkaConsumer.Brokers,
		cfg.RawTopic,
		"preprocessor-manual",
		agg,
		log,
	)

	g.Go(func() error {
		return manualConsumer.Run(ctx)
	})

	// === Consumer стаканов (OrderBook) ===
	orderbookWriter := timescaledb.NewOrderBookWriter(tsWriter.Pool(), log)
	orderbookHandler := kafka.NewOrderBookHandler(orderbookWriter, log)

	orderbookTopic := cfg.OrderBookTopic
	if orderbookTopic == "" {
		orderbookTopic = "marketdata.orderbook"
	}
	log.WithContext(ctx).Info(" Запускаем обработку стаканов заявок",
		zap.String("topic", orderbookTopic),
	)

	g.Go(func() error {
		return kcons.Consume(ctx, []string{orderbookTopic}, func(msg *kafkamsg.Message) error {
			return orderbookHandler.Handle(ctx, msg)
		})
	})

	// === Ждём завершения всех goroutines ===
	if err := g.Wait(); err != nil {
		if ctx.Err() == context.Canceled {
			log.WithContext(ctx).Info(" Preprocessor завершился по Cancel")
			return nil
		}
		log.WithContext(ctx).Error(" Preprocessor аварийно завершился", zap.Error(err))
		return fmt.Errorf("preprocessor exited with error: %w", err)
	}

	log.WithContext(ctx).Info(" Preprocessor завершён успешно")
	return nil
}

// shutdownSafe безопасно завершает компонент с логированием
func shutdownSafe(ctx context.Context, name string, fn func(context.Context) error, log *logger.Logger) {
	log.WithContext(ctx).Info(name + ": shutting down")
	if err := fn(ctx); err != nil {
		log.WithContext(ctx).Error(name+" shutdown failed", zap.Error(err))
	} else {
		log.WithContext(ctx).Info(name + ": shutdown complete")
	}
}
