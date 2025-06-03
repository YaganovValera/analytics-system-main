package timescaledb

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/YaganovValera/analytics-system/common/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	marketdatapb "github.com/YaganovValera/analytics-system/proto/gen/go/v1/marketdata"
)

// OrderBookWriter сохраняет снапшоты стакана в TimescaleDB.
type OrderBookWriter struct {
	db  *pgxpool.Pool
	log *logger.Logger
}

func NewOrderBookWriter(db *pgxpool.Pool, log *logger.Logger) *OrderBookWriter {
	return &OrderBookWriter{
		db:  db,
		log: log.Named("orderbook-writer"),
	}
}

// Insert сериализует bids/asks и записывает snapshot в БД.
func (w *OrderBookWriter) Insert(ctx context.Context, snap *marketdatapb.OrderBookSnapshot) error {
	bids, err := json.Marshal(snap.Bids)
	if err != nil {
		return fmt.Errorf("marshal bids: %w", err)
	}
	asks, err := json.Marshal(snap.Asks)
	if err != nil {
		return fmt.Errorf("marshal asks: %w", err)
	}

	const q = `
		INSERT INTO orderbook_snapshots (symbol, timestamp, bids, asks)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (symbol, timestamp) DO NOTHING
	`
	_, err = w.db.Exec(ctx, q, snap.Symbol, snap.Timestamp.AsTime().UTC(), bids, asks)
	if err != nil {
		w.log.WithContext(ctx).Error("insert failed", zap.Error(err))
		return fmt.Errorf("orderbook insert: %w", err)
	}

	// w.log.WithContext(ctx).Debug("inserted orderbook snapshot",
	// 	zap.String("symbol", snap.Symbol),
	// 	zap.Int("bids", len(snap.Bids)),
	// 	zap.Int("asks", len(snap.Asks)),
	// )
	return nil
}
