// src/types/candle.ts

export type Interval =
  | '1m'
  | '5m'
  | '15m'
  | '1h'
  | '4h'
  | '1d';

export interface ProtoTimestamp {
  seconds: number;
  nanos?: number;
}

export interface Candle {
  symbol: string;
  open_time: ProtoTimestamp;
  close_time: ProtoTimestamp;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
}

export interface CSVParsedCandle {
  symbol: string;
  open_time: Date;
  close_time: Date;
  open: number;
  high: number;
  low: number;
  close: number;
  volume: number;
}

export interface Pagination {
  page_size: number;
  page_token?: string;
}

export interface CandleQuery {
  symbol: string;
  interval: Interval;
  start: string;
  end: string;
  pagination: Pagination;
}

export interface CandleResponse {
  candles: Candle[];
  next_page_token?: string;
}
