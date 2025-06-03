-- preprocessor/migrations/timescaledb/0002_create_orderbook_snapshots.sql

CREATE TABLE IF NOT EXISTS orderbook_snapshots (
  symbol TEXT NOT NULL,
  timestamp TIMESTAMPTZ NOT NULL,
  bids JSONB NOT NULL,
  asks JSONB NOT NULL,
  PRIMARY KEY (symbol, timestamp)
);

-- Индекс для ускоренного поиска по времени и символу
CREATE INDEX IF NOT EXISTS idx_orderbook_symbol_ts ON orderbook_snapshots (symbol, timestamp DESC);
