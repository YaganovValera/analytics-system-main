package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	once              sync.Once
	EventsTotal       prometheus.Counter
	UnsupportedEvents prometheus.Counter
	ParseErrors       prometheus.Counter
	SerializeErrors   prometheus.Counter
	PublishErrors     prometheus.Counter
	PublishLatency    prometheus.Histogram
)

// Register инициализирует и регистрирует все метрики процессора ровно один раз.
func Register(r prometheus.Registerer) {
	once.Do(func() {
		if r == nil {
			r = prometheus.DefaultRegisterer
		}

		EventsTotal = prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "collector", Subsystem: "processor", Name: "events_total",
			Help: "Total number of raw events processed",
		})
		UnsupportedEvents = prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "collector", Subsystem: "processor", Name: "unsupported_events_total",
			Help: "Number of raw events skipped due to unsupported type",
		})
		ParseErrors = prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "collector", Subsystem: "processor", Name: "parse_errors_total",
			Help: "Number of JSON parse errors in raw events",
		})
		SerializeErrors = prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "collector", Subsystem: "processor", Name: "serialize_errors_total",
			Help: "Number of Protobuf serialization errors",
		})
		PublishErrors = prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "collector", Subsystem: "processor", Name: "publish_errors_total",
			Help: "Number of errors publishing messages to Kafka",
		})
		PublishLatency = prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: "collector", Subsystem: "processor", Name: "publish_latency_seconds",
			Help:    "Latency from WS ingestion to Kafka publish (seconds)",
			Buckets: prometheus.DefBuckets,
		})

		for _, c := range []prometheus.Collector{
			EventsTotal, UnsupportedEvents, ParseErrors, SerializeErrors, PublishErrors, PublishLatency,
		} {
			if err := r.Register(c); err != nil {
				if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
					panic(err)
				}
			}
		}
	})
}
