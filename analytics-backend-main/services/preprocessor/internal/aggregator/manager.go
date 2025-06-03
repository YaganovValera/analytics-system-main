// preprocessor/internal/aggregator/manager.go
package aggregator

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/YaganovValera/analytics-system/common/interval"
	"github.com/YaganovValera/analytics-system/common/logger"
	marketdatapb "github.com/YaganovValera/analytics-system/proto/gen/go/v1/marketdata"
	"github.com/YaganovValera/analytics-system/services/preprocessor/internal/metrics"

	"go.uber.org/zap"
)

// Manager реализует Aggregator.
type Manager struct {
	mu            sync.RWMutex
	buckets       map[string]map[string]*candleState
	intervals     []string
	sink          FlushSink
	storage       PartialBarStorage
	log           *logger.Logger
	now           func() time.Time
	flushInterval time.Duration

	ctx    context.Context
	cancel context.CancelFunc
}

// NewManager создаёт и валидацирует Manager.
func NewManager(intervals []string, sink FlushSink, store PartialBarStorage, log *logger.Logger) (*Manager, error) {
	m := &Manager{
		buckets:       make(map[string]map[string]*candleState),
		intervals:     intervals,
		sink:          sink,
		storage:       store,
		log:           log.Named("aggregator"),
		now:           time.Now,
		flushInterval: time.Second,
	}
	m.ctx, m.cancel = context.WithCancel(context.Background())

	for _, iv := range intervals {
		if _, err := interval.Duration(interval.Interval(iv)); err != nil {
			return nil, fmt.Errorf("invalid interval %q: %w", iv, err)
		}
		m.buckets[iv] = make(map[string]*candleState)
	}

	go m.flushLoop()
	return m, nil
}

func (m *Manager) Process(ctx context.Context, data *marketdatapb.MarketData) error {
	symbol := data.Symbol
	price := data.Price
	volume := data.Volume
	ts := data.Timestamp.AsTime()
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, iv := range m.intervals {
		bucket := m.buckets[iv]
		state, ok := bucket[symbol]

		if !ok {
			existing, err := m.storage.LoadAt(ctx, symbol, iv, ts)
			if err != nil {
				metrics.RestoreErrorsTotal.WithLabelValues(iv).Inc()
				m.log.WithContext(ctx).Warn("restore from redis failed",
					zap.String("symbol", symbol), zap.String("interval", iv), zap.Error(err))
			}
			if existing != nil {
				state = &candleState{Candle: existing, UpdatedAt: ts}
			} else {
				state, err = newCandleState(symbol, iv, ts, price, volume)
				if err != nil {
					m.log.WithContext(ctx).Error("newCandleState failed", zap.Error(err))
					continue
				}
			}
			bucket[symbol] = state
			metrics.ProcessedTotal.WithLabelValues(iv).Inc()
			_ = m.storage.Save(ctx, state.Candle)
			continue
		}

		if state.shouldFlush(m.now()) {
			var err error
			if err = m.flushOne(ctx, iv, symbol, state); err != nil {
				m.log.WithContext(ctx).Error("flush failed",
					zap.String("interval", iv), zap.String("symbol", symbol), zap.Error(err))
			}
			state, err = newCandleState(symbol, iv, ts, price, volume)
			if err != nil {
				m.log.WithContext(ctx).Error("newCandleState after flush failed", zap.Error(err))
				continue
			}
			bucket[symbol] = state
		} else {
			state.update(ts, price, volume)
		}

		metrics.ProcessedTotal.WithLabelValues(iv).Inc()
		_ = m.storage.Save(ctx, state.Candle)
	}

	return nil
}

func (m *Manager) flushOne(ctx context.Context, iv, symbol string, state *candleState) error {
	state.Candle.Complete = true
	if err := m.sink.FlushCandle(ctx, state.Candle); err != nil {
		return err
	}
	if err := m.storage.DeleteAt(ctx, symbol, iv, state.Candle.Start); err != nil {
		m.log.WithContext(ctx).Warn("delete from redis failed", zap.Error(err))
	}

	metrics.FlushedTotal.WithLabelValues(iv).Inc()
	latency := m.now().Sub(state.UpdatedAt).Seconds()
	metrics.FlushLatency.WithLabelValues(iv).Observe(latency)
	metrics.LastFlushTimestamp.WithLabelValues(iv).Set(float64(m.now().Unix()))

	return nil
}

func (m *Manager) FlushAll(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errs []error
	for iv, bucket := range m.buckets {
		for symbol, state := range bucket {
			if err := m.flushOne(ctx, iv, symbol, state); err != nil {
				errs = append(errs, err)
				m.log.WithContext(ctx).Error("flushAll failed",
					zap.String("interval", iv), zap.String("symbol", symbol), zap.Error(err))
			}
			delete(bucket, symbol)
		}
	}
	return errors.Join(errs...)
}

func (m *Manager) Close() error {
	m.cancel()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return m.FlushAll(ctx)
}

func (m *Manager) flushLoop() {
	ticker := time.NewTicker(m.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.flushExpired()
		case <-m.ctx.Done():
			m.log.Info("flushLoop: exiting on context cancel")
			return
		}
	}
}

func (m *Manager) flushExpired() {
	now := m.now()
	m.mu.Lock()
	defer m.mu.Unlock()

	for iv, bucket := range m.buckets {
		for symbol, state := range bucket {
			if state.shouldFlush(now) {
				if err := m.flushOne(m.ctx, iv, symbol, state); err != nil {
					m.log.WithContext(m.ctx).Error("flushExpired failed",
						zap.String("interval", iv), zap.String("symbol", symbol), zap.Error(err))
				}
				delete(bucket, symbol)
			}
		}
	}
}
