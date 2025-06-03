// query-service/internal/storage/timescaledb/orderbook.go

package timescaledb

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/YaganovValera/analytics-system/services/query-service/internal/metrics"

	marketdatapb "github.com/YaganovValera/analytics-system/proto/gen/go/v1/marketdata"

	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// GetOrderBook извлекает снапшоты стаканов за указанный период.
func (r *repo) GetOrderBook(ctx context.Context, symbol string, start, end time.Time, limit int, pageToken string) ([]*marketdatapb.OrderBookSnapshot, string, error) {
	startTime := time.Now()
	source := "timescaledb"

	var afterTs time.Time
	var err error

	if pageToken != "" {
		afterTs, err = time.Parse(time.RFC3339Nano, pageToken)
		if err != nil {
			metrics.QueryErrorsTotal.WithLabelValues("invalid_page_token", source).Inc()
			r.log.WithContext(ctx).Warn("invalid page token", zap.String("token", pageToken), zap.Error(err))
			return nil, "", fmt.Errorf("invalid page token: %w", err)
		}
	} else {
		afterTs = start
	}

	const query = `
		SELECT timestamp, bids, asks
		FROM orderbook_snapshots
		WHERE symbol = $1 AND timestamp > $2 AND timestamp <= $3
		ORDER BY timestamp ASC
		LIMIT $4
	`

	rows, err := r.db.Query(ctx, query, symbol, afterTs, end, limit)
	if err != nil {
		metrics.QueryErrorsTotal.WithLabelValues("query", source).Inc()
		r.log.WithContext(ctx).Error("orderbook query failed", zap.Error(err))
		return nil, "", fmt.Errorf("query orderbook_snapshots: %w", err)
	}
	defer rows.Close()

	var results []*marketdatapb.OrderBookSnapshot
	var nextPageToken string

	for rows.Next() {
		var ts time.Time
		var bidsData, asksData []byte

		if err := rows.Scan(&ts, &bidsData, &asksData); err != nil {
			metrics.QueryErrorsTotal.WithLabelValues("scan", source).Inc()
			r.log.WithContext(ctx).Error("scan failed", zap.Error(err))
			return nil, "", fmt.Errorf("scan orderbook row: %w", err)
		}

		var bids, asks []*marketdatapb.OrderBookLevel
		if err := json.Unmarshal(bidsData, &bids); err != nil {
			metrics.QueryErrorsTotal.WithLabelValues("unmarshal_bids", source).Inc()
			r.log.WithContext(ctx).Error("failed to unmarshal bids", zap.ByteString("data", bidsData), zap.Error(err))
			return nil, "", fmt.Errorf("unmarshal bids: %w", err)
		}
		if err := json.Unmarshal(asksData, &asks); err != nil {
			metrics.QueryErrorsTotal.WithLabelValues("unmarshal_asks", source).Inc()
			r.log.WithContext(ctx).Error("failed to unmarshal asks", zap.ByteString("data", asksData), zap.Error(err))
			return nil, "", fmt.Errorf("unmarshal asks: %w", err)
		}

		results = append(results, &marketdatapb.OrderBookSnapshot{
			Timestamp: timestamppb.New(ts),
			Symbol:    symbol,
			Bids:      bids,
			Asks:      asks,
		})

		nextPageToken = ts.Format(time.RFC3339Nano)
	}

	metrics.QueryLatency.WithLabelValues(source).Observe(time.Since(startTime).Seconds())
	metrics.QuerySuccessTotal.WithLabelValues(source).Inc()
	metrics.QueryRowsReturned.WithLabelValues(source).Observe(float64(len(results)))

	return results, nextPageToken, nil
}
