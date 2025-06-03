package binance

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/YaganovValera/analytics-system/common/backoff"
	"github.com/YaganovValera/analytics-system/common/logger"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type connector struct {
	cfg         Config
	log         *logger.Logger
	subscribeID uint64

	mu         sync.Mutex
	conn       *websocket.Conn
	cancelPing context.CancelFunc
	closed     atomic.Bool
}

func NewConnector(cfg Config, log *logger.Logger) (Connector, error) {
	cfg.ApplyDefaults()
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &connector{
		cfg: cfg,
		log: log.Named("binance-ws"),
	}, nil
}

func (c *connector) Stream(ctx context.Context) (<-chan RawMessage, error) {
	ch := make(chan RawMessage, c.cfg.BufferSize)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				c.log.Error("panic in run", zap.Any("error", r))
			}
			close(ch)
		}()
		c.run(ctx, ch)
	}()
	return ch, nil
}

func (c *connector) Close() error {
	c.closed.Store(true)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancelPing != nil {
		c.cancelPing()
	}
	if c.conn != nil {
		_ = c.conn.Close()
	}
	return nil
}

func (c *connector) run(ctx context.Context, ch chan<- RawMessage) {
	for {
		if ctx.Err() != nil || c.closed.Load() {
			c.log.Info("ws: stopping run-loop")
			return
		}

		conn, err := c.connect(ctx)
		if err != nil {
			c.log.Error("ws: connect failed", zap.Error(err))
			return
		}
		conn.SetReadLimit(2 * 1024 * 1024) // 2MB

		c.mu.Lock()
		c.conn = conn
		c.cancelPing = c.startPinger(ctx, conn)
		c.mu.Unlock()

		c.log.Info("ws: connected", zap.String("url", c.cfg.URL))

		if err := c.subscribe(ctx, conn); err != nil {
			c.log.Error("ws: subscribe failed", zap.Error(err))
			_ = conn.Close()
			continue
		}
		c.log.Info("ws: subscribed", zap.Strings("streams", c.cfg.Streams))

		if err := c.readLoop(ctx, conn, ch); err != nil {
			c.log.Warn("ws: read-loop error, reconnecting", zap.Error(err))
		}

		c.cancelPing()
		_ = conn.Close()
	}
}

func (c *connector) connect(ctx context.Context) (*websocket.Conn, error) {
	var conn *websocket.Conn
	err := backoff.Execute(ctx, c.cfg.BackoffConfig, func(ctx context.Context) error {
		var err error
		conn, _, err = websocket.DefaultDialer.DialContext(ctx, c.cfg.URL, nil)
		return err
	}, func(ctx context.Context, err error, delay time.Duration, attempt int) {
		c.log.WithContext(ctx).Warn("binance ws connect retry",
			zap.Int("attempt", attempt),
			zap.Duration("delay", delay),
			zap.Error(err),
			zap.String("url", c.cfg.URL),
		)
	})
	return conn, err
}

func (c *connector) startPinger(ctx context.Context, conn *websocket.Conn) context.CancelFunc {
	conn.SetReadDeadline(time.Now().Add(c.cfg.ReadTimeout))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(c.cfg.ReadTimeout))
	})

	pingCtx, cancel := context.WithCancel(ctx)
	ticker := time.NewTicker(c.cfg.ReadTimeout / 3)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-pingCtx.Done():
				return
			case <-ticker.C:
				conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
				if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(1*time.Second)); err != nil {
					c.log.Warn("ws: ping failed", zap.Error(err))
				}
			}
		}
	}()
	return cancel
}

func (c *connector) subscribe(ctx context.Context, conn *websocket.Conn) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	id := atomic.AddUint64(&c.subscribeID, 1)
	req := map[string]interface{}{
		"method": "SUBSCRIBE",
		"params": c.cfg.Streams,
		"id":     id,
	}
	conn.SetWriteDeadline(time.Now().Add(c.cfg.SubscribeTimeout))
	return conn.WriteJSON(req)
}

func (c *connector) readLoop(ctx context.Context, conn *websocket.Conn, out chan<- RawMessage) error {
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		_, rawBytes, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err,
				websocket.CloseNormalClosure,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure,
			) {
				c.log.Info("ws: connection closed", zap.Error(err))
			} else {
				c.log.Warn("ws: read error", zap.Error(err))
			}
			return err
		}

		var wrapper struct {
			Stream string          `json:"stream"`
			Data   json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(rawBytes, &wrapper); err == nil && len(wrapper.Data) > 0 {
			rawBytes = wrapper.Data
		}

		var meta map[string]any
		if err := json.Unmarshal(rawBytes, &meta); err != nil {
			c.log.Warn("ws: failed to parse event metadata", zap.Error(err))
			continue
		}

		var msgType string
		if v, ok := meta["e"]; ok {
			if s, ok := v.(string); ok {
				msgType = s
			} else {
				c.log.Warn("ws: unexpected type for event type", zap.Any("e", v))
				msgType = "unknown"
			}
		} else {
			msgType = "unknown"
		}

		if msgType == "" {
			msgType = "unknown"
			c.log.Warn("ws: unknown event type", zap.ByteString("data", rawBytes))
		}

		select {
		case out <- RawMessage{Data: rawBytes, Type: msgType}:
		default:
			c.log.Warn("ws: buffer full, dropping message", zap.String("type", msgType))
		}
	}
}
