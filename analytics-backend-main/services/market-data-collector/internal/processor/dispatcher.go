package processor

import (
	"context"
	"fmt"

	"github.com/YaganovValera/analytics-system/common/logger"
	"github.com/YaganovValera/analytics-system/services/market-data-collector/internal/metrics"
	"github.com/YaganovValera/analytics-system/services/market-data-collector/pkg/binance"
	"go.uber.org/zap"
)

type Processor interface {
	Process(ctx context.Context, raw binance.RawMessage) error
}

type DispatchRouter struct {
	processors map[string]Processor
	log        *logger.Logger
}

func NewRouter(log *logger.Logger) *DispatchRouter {
	return &DispatchRouter{
		processors: make(map[string]Processor),
		log:        log.Named("router"),
	}
}

func (r *DispatchRouter) Register(eventType string, proc Processor) {
	r.processors[eventType] = proc
	r.log.Info("registered processor",
		zap.String("event_type", eventType),
		zap.String("impl", fmt.Sprintf("%T", proc)),
	)
}

func (r *DispatchRouter) Run(ctx context.Context, in <-chan binance.RawMessage) error {
	for msg := range in {
		evtType := msg.Type
		proc, ok := r.processors[evtType]
		if !ok {
			metrics.UnsupportedEvents.Inc()
			r.log.WithContext(ctx).Debug("unsupported event type",
				zap.String("event_type", evtType),
			)
			continue
		}

		err := proc.Process(ctx, msg)
		if err != nil {
			r.log.WithContext(ctx).Error("event processing failed",
				zap.String("event_type", evtType),
				zap.Error(err),
			)
			metrics.PublishErrors.Inc()
		}
	}
	return nil
}
