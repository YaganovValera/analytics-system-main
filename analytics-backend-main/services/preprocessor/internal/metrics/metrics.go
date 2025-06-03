// preprocessor/internal/metrics/metrics.go

package metrics

import (
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	once sync.Once

	ProcessedTotal     *prometheus.CounterVec
	FlushedTotal       *prometheus.CounterVec
	FlushLatency       *prometheus.HistogramVec
	RestoreErrorsTotal *prometheus.CounterVec
	LastFlushTimestamp *prometheus.GaugeVec

	OrderbookProcessed   prometheus.Counter
	OrderbookFailed      prometheus.Counter
	InvalidProtoMsgTotal prometheus.Counter

	KafkaPublishedTotal      *prometheus.CounterVec
	KafkaPublishFailedTotal  *prometheus.CounterVec
	RedisSaveFailedTotal     *prometheus.CounterVec
	RedisRestoreSuccessTotal *prometheus.CounterVec
	RedisRestoreFailedTotal  *prometheus.CounterVec
)

func Register(r prometheus.Registerer) {
	once.Do(func() {
		if r == nil {
			r = prometheus.DefaultRegisterer
		}

		ProcessedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "preprocessor", Subsystem: "aggregator", Name: "processed_total",
			Help: "Total MarketData ticks processed",
		}, []string{"interval"})

		FlushedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "preprocessor", Subsystem: "aggregator", Name: "flushed_total",
			Help: "Total finalized candles flushed",
		}, []string{"interval"})

		FlushLatency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "preprocessor", Subsystem: "aggregator", Name: "flush_latency_seconds",
			Help:    "Time between last tick and flush",
			Buckets: prometheus.DefBuckets,
		}, []string{"interval"})

		RestoreErrorsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "preprocessor", Subsystem: "aggregator", Name: "restore_errors_total",
			Help: "Number of failed partial bar restore attempts",
		}, []string{"interval"})

		LastFlushTimestamp = prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "preprocessor", Subsystem: "aggregator", Name: "last_flush_timestamp",
			Help: "Timestamp of last successful flush (unix seconds)",
		}, []string{"interval"})

		OrderbookProcessed = prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "preprocessor", Subsystem: "orderbook", Name: "processed_total",
			Help: "Total orderbook snapshots processed",
		})

		OrderbookFailed = prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "preprocessor", Subsystem: "orderbook", Name: "failed_total",
			Help: "Total orderbook insert failures",
		})

		InvalidProtoMsgTotal = prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: "preprocessor", Subsystem: "kafka", Name: "invalid_proto_total",
			Help: "Total invalid protobuf messages received",
		})

		KafkaPublishedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "preprocessor", Subsystem: "kafka", Name: "published_total",
			Help: "Total candles published to Kafka",
		}, []string{"interval"})

		KafkaPublishFailedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "preprocessor", Subsystem: "kafka", Name: "publish_failed_total",
			Help: "Failed Kafka publish attempts",
		}, []string{"interval"})

		RedisSaveFailedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "preprocessor", Subsystem: "redis", Name: "save_failed_total",
			Help: "Redis save failures",
		}, []string{"interval"})

		RedisRestoreSuccessTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "preprocessor", Subsystem: "redis", Name: "restore_success_total",
			Help: "Redis restore successes",
		}, []string{"interval"})

		RedisRestoreFailedTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "preprocessor", Subsystem: "redis", Name: "restore_failed_total",
			Help: "Redis restore failures",
		}, []string{"interval"})

		collectors := []prometheus.Collector{
			ProcessedTotal,
			FlushedTotal,
			FlushLatency,
			RestoreErrorsTotal,
			LastFlushTimestamp,
			OrderbookProcessed,
			OrderbookFailed,
			InvalidProtoMsgTotal,
			KafkaPublishedTotal,
			KafkaPublishFailedTotal,
			RedisSaveFailedTotal,
			RedisRestoreSuccessTotal,
			RedisRestoreFailedTotal,
		}

		for _, c := range collectors {
			if err := r.Register(c); err != nil {
				if _, ok := err.(prometheus.AlreadyRegisteredError); !ok {
					panic(err)
				}
			}
		}
	})
}
