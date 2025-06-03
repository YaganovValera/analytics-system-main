// preprocessor/internal/storage/timescaledb/timescaledb.go

package timescaledb

import (
	"context"
	"fmt"
	"time"

	"github.com/YaganovValera/analytics-system/common/logger"
	"github.com/YaganovValera/analytics-system/common/serviceid"
	"github.com/YaganovValera/analytics-system/services/preprocessor/internal/aggregator"

	migr "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// ApplyMigrations применяет все миграции в директории.
func ApplyMigrations(cfg Config, log *logger.Logger) error {
	serviceid.InitServiceName(cfg.MigrationsDir)

	m, err := migr.New(
		"file://"+cfg.MigrationsDir,
		cfg.DSN,
	)
	if err != nil {
		return fmt.Errorf("timescaledb: migrate init: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migr.ErrNoChange {
		return fmt.Errorf("timescaledb: migrate up: %w", err)
	}

	log.Info("timescaledb: migrations applied successfully")
	return nil
}

// TimescaleWriter умеет писать свечи в БД.
type TimescaleWriter struct {
	db  *pgxpool.Pool
	log *logger.Logger
}

func NewTimescaleWriter(cfg Config, log *logger.Logger) (*TimescaleWriter, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pgxCfg, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("timescaledb: parse dsn: %w", err)
	}

	db, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		return nil, fmt.Errorf("timescaledb: connect: %w", err)
	}

	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("timescaledb: ping failed: %w", err)
	}

	return &TimescaleWriter{
		db:  db,
		log: log.Named("timescaledb"),
	}, nil
}

// FlushCandle вставляет или игнорирует дубликат.
func (w *TimescaleWriter) FlushCandle(ctx context.Context, c *aggregator.Candle) error {
	const query = `INSERT INTO candles (
		time, symbol, interval, open, high, low, close, volume
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT (symbol, interval, time) DO NOTHING`

	_, err := w.db.Exec(ctx, query,
		c.Start.UTC(),
		c.Symbol,
		c.Interval,
		c.Open,
		c.High,
		c.Low,
		c.Close,
		c.Volume,
	)
	if err != nil {
		w.log.WithContext(ctx).
			Error("timescaledb insert failed",
				zap.String("symbol", c.Symbol),
				zap.String("interval", c.Interval),
				zap.Error(err))
		return fmt.Errorf("timescaledb insert: %w", err)
	}

	w.log.WithContext(ctx).Debug("inserted candle into timescaledb",
		zap.String("symbol", c.Symbol),
		zap.String("interval", c.Interval),
		zap.Time("start", c.Start),
	)
	return nil
}

func (w *TimescaleWriter) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return w.db.Ping(ctx)
}

func (w *TimescaleWriter) Close() {
	w.db.Close()
}

func (w *TimescaleWriter) Pool() *pgxpool.Pool {
	return w.db
}
