import type { Candle } from './candle';

export interface AnalyticsResponse {
  symbol: string;
  interval: string;
  start: string;
  end: string;
  count: number;
  analytics: {
    avg_close: number;
    sum_volume: number;
    price_change: number;
    volatility: number;
    up_count: number;
    down_count: number;
    up_ratio: number;
    down_ratio: number;
    avg_body_size: number;
    avg_upper_wick: number;
    avg_lower_wick: number;
    max_gap_up: number;
    max_gap_down: number;
    bullish_streak: number;
    bearish_streak: number;
    price_range_percent: number;
    dominant_hour: number;
    max_candle: Candle;
    min_candle: Candle;
    most_volatile_candle: Candle;
    most_volume_candle: Candle;
    max_gap_up_candle: Candle;
    max_gap_down_candle: Candle;
  };
}
