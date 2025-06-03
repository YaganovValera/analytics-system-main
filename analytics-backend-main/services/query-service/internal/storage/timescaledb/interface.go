// query-service/internal/storage/timescaledb/interface.go
package timescaledb

import (
	"context"
	"time"

	marketdatapb "github.com/YaganovValera/analytics-system/proto/gen/go/v1/marketdata"
)

type Repository interface {
	ExecuteSQL(ctx context.Context, query string, params map[string]string) (columns []string, rows [][]string, err error)
	Ping(ctx context.Context) error
	Close()
	GetOrderBook(ctx context.Context, symbol string, start, end time.Time, limit int, pageToken string) ([]*marketdatapb.OrderBookSnapshot, string, error)
}
