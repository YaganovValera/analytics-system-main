package binance

import (
	"context"

	"github.com/YaganovValera/analytics-system/common/logger"
	"github.com/YaganovValera/analytics-system/services/market-data-collector/pkg/binance"
	"go.uber.org/zap"
)

// StreamWithMetrics wraps the raw connector with metrics and safe buffering.
func StreamWithMetrics(ctx context.Context, conn binance.Connector) (<-chan binance.RawMessage, error) {
	log := logger.FromContext(ctx).Named("binance.client")

	stream, err := conn.Stream(ctx)
	if err != nil {
		IncError("connect")
		log.Error("failed to open Binance stream", zap.Error(err))
		return nil, err
	}
	IncConnect("ok")
	log.Info("Binance stream started", zap.Int("buffer", cap(stream)))

	out := make(chan binance.RawMessage, cap(stream))

	go func() {
		defer close(out)
		for msg := range stream {
			IncMessage(msg.Type)

			select {
			case out <- msg:
				// ok
			default:
				IncDrop(msg.Type)
				log.Warn("message dropped due to full buffer", zap.String("type", msg.Type))
			}
		}
		log.Info("Binance stream ended")
	}()

	return out, nil
}
