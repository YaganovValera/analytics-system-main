-- preprocessor/migrations/timescaledb/0001_create_candles.up.sql
CREATE TABLE IF NOT EXISTS candles (
  time TIMESTAMPTZ NOT NULL,
  symbol TEXT NOT NULL,
  interval TEXT NOT NULL,
  open DOUBLE PRECISION NOT NULL,
  high DOUBLE PRECISION NOT NULL,
  low DOUBLE PRECISION NOT NULL,
  close DOUBLE PRECISION NOT NULL,
  volume DOUBLE PRECISION NOT NULL,
  PRIMARY KEY (symbol, interval, time)
);

CREATE INDEX IF NOT EXISTS idx_candles_symbol_interval ON candles (symbol, interval);

SELECT create_hypertable('candles', 'time', if_not_exists => TRUE);
