package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/YaganovValera/analytics-system/common/backoff"
	httpserver "github.com/YaganovValera/analytics-system/common/httpserver"
	producer "github.com/YaganovValera/analytics-system/common/kafka/producer"
	"github.com/YaganovValera/analytics-system/common/logger"
	"github.com/YaganovValera/analytics-system/common/serviceid"

	transportBinance "github.com/YaganovValera/analytics-system/services/market-data-collector/internal/transport/binance"
	pkgBinance "github.com/YaganovValera/analytics-system/services/market-data-collector/pkg/binance"

	"github.com/YaganovValera/analytics-system/services/market-data-collector/internal/config"
	"github.com/YaganovValera/analytics-system/services/market-data-collector/internal/metrics"
	"github.com/YaganovValera/analytics-system/services/market-data-collector/internal/processor"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func Run(ctx context.Context, cfg *config.Config, log *logger.Logger) error {
	serviceid.InitServiceName(cfg.ServiceName)
	metrics.Register(nil)
	transportBinance.RegisterMetrics(nil)

	log.Info("service initialization started")

	// === Binance Connector
	binanceCfg := pkgBinance.Config{
		URL:              cfg.Binance.URL,
		Streams:          cfg.Binance.Streams,
		BufferSize:       cfg.Binance.BufferSize,
		ReadTimeout:      cfg.Binance.ReadTimeout,
		SubscribeTimeout: cfg.Binance.SubscribeTimeout,
		BackoffConfig:    cfg.Binance.BackoffConfig,
	}
	wsConn, err := pkgBinance.NewConnector(binanceCfg, log)
	if err != nil {
		return fmt.Errorf("binance connector init: %w", err)
	}
	defer shutdownSafe(ctx, "ws-connector", wsConn.Close, log)

	log.Info("binance websocket connector initialized",
		zap.String("url", cfg.Binance.URL),
		zap.Strings("streams", cfg.Binance.Streams),
	)

	wsManager := transportBinance.NewWSManager(wsConn)
	defer shutdownSafe(ctx, "ws-manager", func() error {
		wsManager.Stop()
		return nil
	}, log)

	// === Kafka Producer
	kafkaProd, err := producer.New(ctx, producer.Config{
		Brokers:        cfg.Kafka.Brokers,
		RequiredAcks:   cfg.Kafka.RequiredAcks,
		Timeout:        cfg.Kafka.Timeout,
		Compression:    cfg.Kafka.Compression,
		FlushFrequency: cfg.Kafka.FlushFrequency,
		FlushMessages:  cfg.Kafka.FlushMessages,
		Backoff:        cfg.Kafka.Backoff,
	}, log)
	if err != nil {
		return fmt.Errorf("kafka producer init: %w", err)
	}
	defer shutdownSafe(ctx, "kafka-producer", kafkaProd.Close, log)

	log.Info("kafka producer initialized", zap.Strings("brokers", cfg.Kafka.Brokers))

	// === Processors
	tradeProc := processor.NewTradeProcessor(kafkaProd, cfg.Kafka.RawTopic, log)
	depthProc := processor.NewDepthProcessor(kafkaProd, cfg.Kafka.OrderBookTopic, log)

	// === HTTP Server
	readiness := func() error { return kafkaProd.Ping(ctx) }
	httpSrv, err := httpserver.New(
		cfg.HTTP,
		readiness,
		log,
		nil,
		httpserver.RecoverMiddleware,
		httpserver.CORSMiddleware(),
	)
	if err != nil {
		return fmt.Errorf("httpserver init: %w", err)
	}

	log.Info("http server initialized", zap.Int("port", cfg.HTTP.Port))

	g, ctx := errgroup.WithContext(ctx)

	// HTTP server
	g.Go(func() error {
		log.Info("starting http server")
		return httpSrv.Run(ctx)
	})

	// WebSocket loop
	g.Go(func() error {
		for {
			if ctx.Err() != nil {
				log.Info("context canceled, stopping ws loop")
				wsManager.Stop()
				return ctx.Err()
			}

			var msgCh <-chan pkgBinance.RawMessage
			var cancel context.CancelFunc

			err := backoff.Execute(ctx, cfg.Binance.BackoffConfig,
				func(ctx context.Context) error {
					ch, cancelFn, err := wsManager.Start(ctx)
					if err == nil {
						msgCh = ch
						cancel = cancelFn
					}
					return err
				},
				func(ctx context.Context, err error, delay time.Duration, attempt int) {
					log.WithContext(ctx).Warn("ws reconnect attempt",
						zap.Int("attempt", attempt),
						zap.Duration("delay", delay),
						zap.Error(err),
					)
				},
			)
			if err != nil {
				wsManager.Stop()
				return fmt.Errorf("ws connect failed: %w", err)
			}

			log.Info("websocket connected, starting router")

			router := processor.NewRouter(log.Named("router"))
			router.Register("trade", tradeProc)
			router.Register("depthUpdate", depthProc)

			log.Info("event processors registered",
				zap.String("trade_topic", cfg.Kafka.RawTopic),
				zap.String("depth_topic", cfg.Kafka.OrderBookTopic),
			)

			if err := router.Run(ctx, msgCh); err != nil {
				log.WithContext(ctx).Error("router exited", zap.Error(err))
			} else {
				log.Info("router stopped gracefully")
			}

			if cancel != nil {
				cancel()
			}

			log.Info("attempting websocket reconnect")
		}
	})

	if err := g.Wait(); err != nil {
		if errors.Is(err, context.Canceled) {
			log.WithContext(ctx).Info("collector stopped by context")
			return nil
		}
		log.WithContext(ctx).Error("collector exited with error", zap.Error(err))
		return err
	}

	log.Info("collector shut down cleanly")
	return nil
}

func shutdownSafe(ctx context.Context, name string, fn func() error, log *logger.Logger) {
	log.WithContext(ctx).Info(fmt.Sprintf("%s: shutting down", name))
	if err := fn(); err != nil {
		log.WithContext(ctx).Error(fmt.Sprintf("%s shutdown error", name), zap.Error(err))
	} else {
		log.WithContext(ctx).Info(fmt.Sprintf("%s: shutdown complete", name))
	}
}
