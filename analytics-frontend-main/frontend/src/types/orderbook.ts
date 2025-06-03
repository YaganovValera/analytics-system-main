// src/types/orderbook.ts

export interface OrderBookLevel {
    price: number;
    quantity: number;
}

export interface OrderBookSnapshot {
    timestamp: { seconds: number };
    symbol: string;
    bids: OrderBookLevel[];
    asks: OrderBookLevel[];
}

export interface OrderBookAnalysis {
    avg_spread_percent: number;
    avg_bid_volume_top10: number;
    avg_ask_volume_top10: number;
    imbalance_start: number;
    imbalance_end: number;
    bid_slope: number;
    ask_slope: number;
    max_bid_wall_price: number;
    max_bid_wall_volume: number;
    max_ask_wall_price: number;
    max_ask_wall_volume: number;
}
  