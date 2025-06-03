// preprocessor/internal/config/config.go
package config

import (
	"fmt"

	commoncfg "github.com/YaganovValera/analytics-system/common/config"
	commonhttp "github.com/YaganovValera/analytics-system/common/httpserver"
	consumercfg "github.com/YaganovValera/analytics-system/common/kafka/consumer"
	producercfg "github.com/YaganovValera/analytics-system/common/kafka/producer"
	commonlogger "github.com/YaganovValera/analytics-system/common/logger"
	commonredis "github.com/YaganovValera/analytics-system/common/redis"
	timescaledb "github.com/YaganovValera/analytics-system/services/preprocessor/internal/storage/timescaledb"
)

// Config хранит параметры приложения preprocessor.
type Config struct {
	ServiceName    string `mapstructure:"service_name"`
	ServiceVersion string `mapstructure:"service_version"`

	KafkaConsumer     consumercfg.Config `mapstructure:"kafka_consumer"`
	KafkaProducer     producercfg.Config `mapstructure:"kafka_producer"`
	RawTopic          string             `mapstructure:"raw_topic"`
	OutputTopicPrefix string             `mapstructure:"output_topic_prefix"`
	OrderBookTopic    string             `mapstructure:"orderbook_topic"`

	Redis     commonredis.Config `mapstructure:"redis"`
	Timescale timescaledb.Config `mapstructure:"timescaledb"`

	Logging   commonlogger.Config `mapstructure:"logging"`
	HTTP      commonhttp.Config   `mapstructure:"http"`
	Intervals []string            `mapstructure:"intervals"`
}

// Load читает конфигурацию из файла и среды окружения.
func Load(path string) (*Config, error) {
	var cfg Config
	if err := commoncfg.Load(commoncfg.Options{
		Path:      path,
		EnvPrefix: "PREPROCESSOR",
		Out:       &cfg,
		Defaults: map[string]interface{}{
			"service_name":    "preprocessor",
			"service_version": "v1.0.0",

			"kafka_consumer.brokers":  []string{"kafka:9092"},
			"kafka_consumer.version":  "2.8.0",
			"kafka_consumer.group_id": "preprocessor",
			"kafka_consumer.backoff": map[string]interface{}{
				"initial_interval": "1s",
				"max_interval":     "30s",
				"max_elapsed_time": "5m",
			},

			"kafka_producer.brokers":         []string{"kafka:9092"},
			"kafka_producer.required_acks":   "all",
			"kafka_producer.timeout":         "15s",
			"kafka_producer.compression":     "none",
			"kafka_producer.flush_frequency": "0s",
			"kafka_producer.flush_messages":  0,
			"kafka_producer.backoff": map[string]interface{}{
				"initial_interval": "1s",
				"max_interval":     "30s",
				"max_elapsed_time": "5m",
			},

			"raw_topic":           "marketdata.raw",
			"output_topic_prefix": "candles",
			"orderbook_topic":     "marketdata.orderbook",

			"redis.addr":         "redis:6379",
			"redis.password":     "",
			"redis.db":           0,
			"redis.service_name": "preprocessor",
			"redis.backoff": map[string]interface{}{
				"initial_interval": "1s",
				"max_interval":     "5s",
				"max_elapsed_time": "30s",
			},

			"timescaledb.dsn":            "postgres://user:pass@timescaledb:5432/analytics?sslmode=disable",
			"timescaledb.migrations_dir": "/app/migrations/timescaledb",

			"logging.level":    "info",
			"logging.dev_mode": false,
			"logging.format":   "console",

			"http.port":             8081,
			"http.read_timeout":     "10s",
			"http.write_timeout":    "15s",
			"http.idle_timeout":     "60s",
			"http.shutdown_timeout": "5s",
			"http.metrics_path":     "/metrics",
			"http.healthz_path":     "/healthz",
			"http.readyz_path":      "/readyz",

			"intervals": []string{"1m", "5m", "15m", "1h", "4h", "1d"},
		},
	}); err != nil {
		return nil, fmt.Errorf("config load failed: %w", err)
	}

	cfg.KafkaConsumer.ApplyDefaults()
	if err := cfg.KafkaConsumer.Validate(); err != nil {
		return nil, fmt.Errorf("kafka_consumer config invalid: %w", err)
	}
	cfg.KafkaProducer.ApplyDefaults()
	if err := cfg.KafkaProducer.Validate(); err != nil {
		return nil, fmt.Errorf("kafka_producer config invalid: %w", err)
	}
	cfg.Redis.ApplyDefaults()
	if err := cfg.Redis.Validate(); err != nil {
		return nil, fmt.Errorf("redis config invalid: %w", err)
	}
	cfg.Timescale.ApplyDefaults()
	if err := cfg.Timescale.Validate(); err != nil {
		return nil, fmt.Errorf("timescaledb config invalid: %w", err)
	}
	cfg.Logging.ApplyDefaults()
	if err := cfg.Logging.Validate(); err != nil {
		return nil, fmt.Errorf("logging config invalid: %w", err)
	}
	cfg.HTTP.ApplyDefaults()
	if err := cfg.HTTP.Validate(); err != nil {
		return nil, fmt.Errorf("http config invalid: %w", err)
	}

	if len(cfg.Intervals) == 0 {
		return nil, fmt.Errorf("intervals must not be empty")
	}

	return &cfg, nil
}
