package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
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

const EventTypeDepth = "depthUpdate"

type depthProcessor struct {
	producer kafka.Producer
	topic    string
	log      *logger.Logger
}

func NewDepthProcessor(p kafka.Producer, topic string, log *logger.Logger) Processor {
	return &depthProcessor{producer: p, topic: topic, log: log.Named("depth")}
}

func (dp *depthProcessor) Process(ctx context.Context, raw binance.RawMessage) error {
	if raw.Type != EventTypeDepth {
		return nil
	}

	metrics.EventsTotal.Inc()

	var evt struct {
		EventType string     `json:"e"`
		EventTime int64      `json:"E"`
		Symbol    string     `json:"s"`
		Bids      [][]string `json:"b"`
		Asks      [][]string `json:"a"`
	}

	if err := json.Unmarshal(raw.Data, &evt); err != nil {
		metrics.ParseErrors.Inc()
		dp.log.WithContext(ctx).Error("unmarshal depth failed", zap.Error(err))
		return nil
	}

	if evt.EventTime <= 0 || evt.Symbol == "" {
		dp.log.WithContext(ctx).Warn("invalid depth: missing required fields",
			zap.Int64("event_time", evt.EventTime),
			zap.String("symbol", evt.Symbol),
		)
		return nil
	}

	convert := func(pair []string) (*marketdata.OrderBookLevel, error) {
		if len(pair) < 2 {
			return nil, fmt.Errorf("malformed level: %v", strings.Join(pair, ","))
		}
		price, err := strconv.ParseFloat(pair[0], 64)
		if err != nil {
			return nil, err
		}
		qty, err := strconv.ParseFloat(pair[1], 64)
		if err != nil {
			return nil, err
		}
		return &marketdata.OrderBookLevel{Price: price, Quantity: qty}, nil
	}

	bids := make([]*marketdata.OrderBookLevel, 0, len(evt.Bids))
	for _, entry := range evt.Bids {
		if lvl, err := convert(entry); err == nil {
			bids = append(bids, lvl)
		} else {
			dp.log.WithContext(ctx).Debug("skipped malformed bid",
				zap.String("symbol", evt.Symbol),
				zap.String("entry", strings.Join(entry, ",")),
				zap.Error(err),
			)
		}
	}

	asks := make([]*marketdata.OrderBookLevel, 0, len(evt.Asks))
	for _, entry := range evt.Asks {
		if lvl, err := convert(entry); err == nil {
			asks = append(asks, lvl)
		} else {
			dp.log.WithContext(ctx).Debug("skipped malformed ask",
				zap.String("symbol", evt.Symbol),
				zap.String("entry", strings.Join(entry, ",")),
				zap.Error(err),
			)
		}
	}

	msg := &marketdata.OrderBookSnapshot{
		Timestamp: timestamppb.New(time.UnixMilli(evt.EventTime)),
		Symbol:    evt.Symbol,
		Bids:      bids,
		Asks:      asks,
	}

	bytes, err := proto.Marshal(msg)
	if err != nil {
		metrics.SerializeErrors.Inc()
		dp.log.WithContext(ctx).Error("proto marshal failed", zap.Error(err))
		return err
	}

	start := time.Now()
	err = dp.producer.Publish(ctx, dp.topic, []byte(evt.Symbol), bytes)
	if err != nil {
		metrics.PublishErrors.Inc()
		dp.log.WithContext(ctx).Error("kafka publish failed",
			zap.String("symbol", evt.Symbol),
			zap.Error(err),
		)
		return err
	}

	metrics.PublishLatency.Observe(time.Since(start).Seconds())

	// dp.log.WithContext(ctx).Debug("depth published",
	// 	zap.String("symbol", evt.Symbol),
	// 	zap.Int("bids", len(bids)),
	// 	zap.Int("asks", len(asks)),
	// )
	return nil
}
