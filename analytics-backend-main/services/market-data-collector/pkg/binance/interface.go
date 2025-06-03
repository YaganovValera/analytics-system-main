package binance

import "context"

type Connector interface {
	Stream(ctx context.Context) (<-chan RawMessage, error)
	Close() error
}

type RawMessage struct {
	Data []byte // JSON event payload
	Type string // e.g. trade, depthUpdate, kline, etc.
}
