// src/api/candles.ts

import api from './axios';
import type { CandleResponse, Interval } from '../types/candle';

// Получение списка символов
export const fetchSymbols = async (): Promise<string[]> => {
  const res = await api.get<{ symbols: string[] }>('symbols');
  return res.data.symbols;
};

interface FetchCandlesParams {
  symbol: string;
  interval: Interval;
  start: string; // ISO8601
  end: string;   // ISO8601
  pageSize?: number;
  pageToken?: string;
}

// Получение свечей с пагинацией
export const fetchCandles = async (params: FetchCandlesParams): Promise<CandleResponse> => {
  const { symbol, interval, start, end, pageSize = 500, pageToken } = params;
  const res = await api.get<CandleResponse>('candles', {
    params: {
      symbol,
      interval,
      start,
      end,
      page_size: pageSize,
      page_token: pageToken,
    },
  });
  return {
    candles: res.data.candles ?? [],
    next_page_token: res.data.next_page_token,
  };
};
