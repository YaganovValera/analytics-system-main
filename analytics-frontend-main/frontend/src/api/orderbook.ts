// src/api/orderbook.ts
import api from './axios';
import type { OrderBookSnapshot, OrderBookAnalysis } from '../types/orderbook';

export interface OrderBookResponse {
  snapshots: OrderBookSnapshot[];
  next_page_token?: string;
  analysis: OrderBookAnalysis;
}

interface FetchOrderBookParams {
  symbol: string;
  start: string;
  end: string;
  pageSize?: number;
  pageToken?: string;
}

export const fetchOrderBook = async (
  params: FetchOrderBookParams
): Promise<OrderBookResponse> => {
  const { symbol, start, end, pageSize = 500, pageToken } = params;

  const res = await api.get<OrderBookResponse>('orderbook', {
    params: {
      symbol,
      start,
      end,
      page_size: pageSize,
      page_token: pageToken,
    },
  });

  return res.data;
};
