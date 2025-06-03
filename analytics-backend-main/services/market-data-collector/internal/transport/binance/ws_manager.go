package binance

import (
	"context"
	"sync"

	"github.com/YaganovValera/analytics-system/common/logger"
	"github.com/YaganovValera/analytics-system/services/market-data-collector/pkg/binance"
	"go.uber.org/zap"
)

// WSManager управляет жизненным циклом WebSocket-потока Binance.
type WSManager struct {
	conn binance.Connector

	mu       sync.Mutex
	cancelFn context.CancelFunc
	log      *logger.Logger
}

func NewWSManager(conn binance.Connector) *WSManager {
	return &WSManager{
		conn: conn,
		log:  logger.NewNamed("binance.ws-manager"),
	}
}

// Start запускает новый поток Binance WebSocket, отменяя предыдущий (если был).
func (w *WSManager) Start(ctx context.Context) (<-chan binance.RawMessage, func(), error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Завершаем предыдущий поток
	if w.cancelFn != nil {
		w.log.Info("stopping previous stream")
		w.cancelFn()
	}

	streamCtx, cancel := context.WithCancel(ctx)
	w.cancelFn = cancel

	msgCh, err := StreamWithMetrics(streamCtx, w.conn)
	if err != nil {
		w.log.Error("failed to start stream", zap.Error(err))
		cancel()
		w.cancelFn = nil
		return nil, nil, err
	}

	w.log.Info("new Binance stream started")
	return msgCh, cancel, nil
}

// Stop завершает активный поток, если он есть.
func (w *WSManager) Stop() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.cancelFn != nil {
		w.log.Info("stream stopped by controller")
		w.cancelFn()
		w.cancelFn = nil
	}
}
