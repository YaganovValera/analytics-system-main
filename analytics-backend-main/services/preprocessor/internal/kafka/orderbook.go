package kafka

import (
	"context"
	"sync"
	"time"

	"github.com/YaganovValera/analytics-system/common/kafka"
	"github.com/YaganovValera/analytics-system/common/logger"
	"github.com/YaganovValera/analytics-system/services/preprocessor/internal/metrics"
	"github.com/YaganovValera/analytics-system/services/preprocessor/internal/storage/timescaledb"

	marketdatapb "github.com/YaganovValera/analytics-system/proto/gen/go/v1/marketdata"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type OrderBookHandler struct {
	writer     *timescaledb.OrderBookWriter
	log        *logger.Logger
	lastWrite  map[string]time.Time
	writeMutex sync.Mutex
}

func NewOrderBookHandler(writer *timescaledb.OrderBookWriter, log *logger.Logger) *OrderBookHandler {
	return &OrderBookHandler{
		writer:    writer,
		log:       log.Named("orderbook-consumer"),
		lastWrite: make(map[string]time.Time),
	}
}

func (h *OrderBookHandler) Handle(ctx context.Context, msg *kafka.Message) error {
	var snapshot marketdatapb.OrderBookSnapshot
	if err := proto.Unmarshal(msg.Value, &snapshot); err != nil {
		h.log.Error("failed to unmarshal OrderBookSnapshot", zap.Error(err))
		metrics.InvalidProtoMsgTotal.Inc()
		return err
	}

	if snapshot.Symbol == "" {
		h.log.Warn("received snapshot with empty symbol", zap.ByteString("key", msg.Key))
		return nil
	}

	now := time.Now().UTC()

	// Rate limit: 1 snapshot/sec per symbol
	h.writeMutex.Lock()
	last, ok := h.lastWrite[snapshot.Symbol]
	if ok && now.Sub(last) < time.Second {
		h.writeMutex.Unlock()
		return nil
	}
	h.lastWrite[snapshot.Symbol] = now
	h.writeMutex.Unlock()

	if err := h.writer.Insert(ctx, &snapshot); err != nil {
		h.log.Error("failed to insert OrderBookSnapshot into DB", zap.Error(err))
		metrics.OrderbookFailed.Inc()
		return err
	}

	// h.log.Debug("orderbook inserted",
	// 	zap.String("symbol", snapshot.Symbol),
	// 	zap.Int("bids", len(snapshot.Bids)),
	// 	zap.Int("asks", len(snapshot.Asks)),
	// )
	metrics.OrderbookProcessed.Inc()
	return nil
}
