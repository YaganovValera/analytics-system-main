package binance

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	once          sync.Once
	wsConnects    *prometheus.CounterVec
	wsErrors      *prometheus.CounterVec
	wsMessages    *prometheus.CounterVec
	wsBufferDrops *prometheus.CounterVec
)

func RegisterMetrics(r prometheus.Registerer) {
	once.Do(func() {
		if r == nil {
			r = prometheus.DefaultRegisterer
		}

		wsConnects = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "collector", Subsystem: "binance", Name: "connects_total",
			Help: "Total WebSocket connection attempts to Binance",
		}, []string{"status"})

		wsErrors = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "collector", Subsystem: "binance", Name: "errors_total",
			Help: "Total categorized WebSocket stream errors",
		}, []string{"type"})

		wsMessages = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "collector", Subsystem: "binance", Name: "messages_total",
			Help: "Total Binance WebSocket messages received by type",
		}, []string{"type"})

		wsBufferDrops = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "collector", Subsystem: "binance", Name: "buffer_drops_total",
			Help: "Messages dropped due to full internal buffer",
		}, []string{"type"})

		for _, c := range []prometheus.Collector{wsConnects, wsErrors, wsMessages, wsBufferDrops} {
			if err := r.Register(c); err != nil {
				if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
					panic(err)
				}
			}
		}
	})
}

func IncConnect(status string)  { wsConnects.WithLabelValues(status).Inc() }
func IncError(errType string)   { wsErrors.WithLabelValues(errType).Inc() }
func IncMessage(msgType string) { wsMessages.WithLabelValues(msgType).Inc() }
func IncDrop(msgType string)    { wsBufferDrops.WithLabelValues(msgType).Inc() }
