package processor

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/YaganovValera/analytics-system/common/kafka"
	"github.com/YaganovValera/analytics-system/common/logger"
	marketdata "github.com/YaganovValera/analytics-system/proto/gen/go/v1/marketdata"
	"github.com/YaganovValera/analytics-system/services/market-data-collector/internal/metrics"
	"github.com/YaganovValera/analytics-system/services/market-data-collector/pkg/binance"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const EventTypeTrade = "trade"

type tradeProcessor struct {
	producer kafka.Producer
	topic    string
	log      *logger.Logger
}

func NewTradeProcessor(p kafka.Producer, topic string, log *logger.Logger) Processor {
	return &tradeProcessor{producer: p, topic: topic, log: log.Named("trade")}
}

func (tp *tradeProcessor) Process(ctx context.Context, raw binance.RawMessage) error {
	if raw.Type != EventTypeTrade {
		return nil
	}

	metrics.EventsTotal.Inc()

	var evt struct {
		EventType string `json:"e"`
		EventTime int64  `json:"E"`
		Symbol    string `json:"s"`
		Price     string `json:"p"`
		Quantity  string `json:"q"`
		TradeID   int64  `json:"t"`
	}

	if err := json.Unmarshal(raw.Data, &evt); err != nil {
		metrics.ParseErrors.Inc()
		tp.log.WithContext(ctx).Error("unmarshal trade failed",
			zap.ByteString("raw", raw.Data),
			zap.Error(err),
		)
		return nil
	}

	if evt.EventTime <= 0 || evt.Symbol == "" {
		tp.log.WithContext(ctx).Warn("invalid trade: missing required fields",
			zap.Int64("event_time", evt.EventTime),
			zap.String("symbol", evt.Symbol),
		)
		return nil
	}

	price, err := strconv.ParseFloat(evt.Price, 64)
	if err != nil {
		metrics.ParseErrors.Inc()
		tp.log.WithContext(ctx).Warn("invalid price format",
			zap.String("price", evt.Price),
			zap.Error(err),
		)
		return nil
	}

	qty, err := strconv.ParseFloat(evt.Quantity, 64)
	if err != nil {
		metrics.ParseErrors.Inc()
		tp.log.WithContext(ctx).Warn("invalid quantity format",
			zap.String("quantity", evt.Quantity),
			zap.Error(err),
		)
		return nil
	}

	msg := &marketdata.MarketData{
		Timestamp: timestamppb.New(time.UnixMilli(evt.EventTime)),
		Symbol:    evt.Symbol,
		Price:     price,
		Volume:    qty,
		BidPrice:  0,
		AskPrice:  0,
		TradeId:   strconv.FormatInt(evt.TradeID, 10),
	}

	bytes, err := proto.Marshal(msg)
	if err != nil {
		metrics.SerializeErrors.Inc()
		tp.log.WithContext(ctx).Error("proto marshal failed", zap.Error(err))
		return err
	}

	start := time.Now()
	err = tp.producer.Publish(ctx, tp.topic, []byte(evt.Symbol), bytes)
	if err != nil {
		metrics.PublishErrors.Inc()
		tp.log.WithContext(ctx).Error("kafka publish failed",
			zap.String("symbol", evt.Symbol),
			zap.Error(err),
		)
		return err
	}

	metrics.PublishLatency.Observe(time.Since(start).Seconds())

	// tp.log.WithContext(ctx).Debug("trade published",
	// 	zap.String("symbol", evt.Symbol),
	// 	zap.Float64("price", price),
	// 	zap.Float64("qty", qty),
	// 	zap.String("trade_id", msg.TradeId),
	// )
	return nil
}
