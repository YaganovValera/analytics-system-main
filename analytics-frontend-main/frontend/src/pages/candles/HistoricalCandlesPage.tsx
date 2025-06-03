// src/pages/candles/HistoricalCandlesPage.tsx

import { useState, useMemo } from 'react';
import CandleForm from '@components/candle/CandleForm';
import CandleTable from '@components/candle/CandleTable';
import { fetchCandles } from '@api/candles';
import type { Interval, Candle } from '../../types/candle';
import './HistoricalCandlesPage.css';

function HistoricalCandlesPage() {
  const [candles, setCandles] = useState<Candle[]>([]);
  const [nextToken, setNextToken] = useState<string | undefined>(undefined);
  const [query, setQuery] = useState<{
    symbol: string;
    interval: Interval;
    start: string;
    end: string;
    total: number;
  } | null>(null);
  const [visibleCount, setVisibleCount] = useState(100);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleQuery = async (params: typeof query) => {
    if (!params) return;
    setQuery(params);
    setCandles([]);
    setNextToken(undefined);
    setVisibleCount(100);
    setError(null);
    setLoading(true);

    try {
      const res = await fetchCandles({
        ...params,
        pageSize: Math.min(params.total, 500),
      });

      const candlesData = res.candles ?? [];
      if (candlesData.length === 0) {
        setCandles([]);
        setNextToken(undefined);
      } else {
        setCandles(deduplicate(candlesData));
        setNextToken(res.next_page_token);
      }
    } catch {
      setError('Ошибка загрузки данных');
    } finally {
      setLoading(false);
    }
  };

  const handleLoadMore = async () => {
    if (!query) return;

    const already = candles.length;
    const remaining = query.total - already;

    if (visibleCount < candles.length) {
      setVisibleCount(Math.min(visibleCount + 100, candles.length));
      return;
    }

    if (!nextToken || remaining <= 0) return;

    setLoading(true);
    try {
      const res = await fetchCandles({
        ...query,
        pageSize: Math.min(500, remaining),
        pageToken: nextToken,
      });

      const newCandles = res.candles ?? [];
      if (newCandles.length > 0) {
        const combined = [...candles, ...newCandles];
        const unique = deduplicate(combined);
        setCandles(unique);
        setVisibleCount(Math.min(visibleCount + 100, unique.length));
        setNextToken(res.next_page_token);
      } else {
        setNextToken(undefined);
      }
    } catch {
      setError('Ошибка при подгрузке данных');
    } finally {
      setLoading(false);
    }
  };

  const handleExportCSV = () => {
    if (!candles.length || !query) return;
  
    const header = 'symbol,open_time,close_time,open,high,low,close,volume';
    const rows = candles.map((c) => [
      c.symbol,
      new Date(c.open_time.seconds * 1000).toISOString(),
      new Date(c.close_time.seconds * 1000).toISOString(),
      c.open,
      c.high,
      c.low,
      c.close,
      c.volume,
    ].join(','));
    
    const content = [header, ...rows].join('\n');
  
    const blob = new Blob([content], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
  
    const a = document.createElement('a');
    a.href = url;
    const now = new Date();
    const timestamp = now.toISOString().replace(/[:T]/g, '-').split('.')[0];
    a.download = `candles_${query.symbol}_${query.interval}_${timestamp}.csv`;
    a.click();
    URL.revokeObjectURL(url);
  };
  


  const stats = useMemo(() => {
    if (candles.length === 0) return null;
    const closeAvg = candles.reduce((sum, c) => sum + c.close, 0) / candles.length;
    const volSum = candles.reduce((sum, c) => sum + c.volume, 0);
    const priceChange = candles[candles.length - 1].close - candles[0].open;
    const maxHigh = Math.max(...candles.map((c) => c.high));
    const minLow = Math.min(...candles.map((c) => c.low));
    const volatility = maxHigh - minLow;

    return { closeAvg, volSum, priceChange, volatility };
  }, [candles]);

  return (
    <div className="candles-layout">
      <div className="candles-main">
        <h2>Исторические свечи</h2>
  
        {stats && (
          <div className="summary-box">
            <h4>Анализ текущей выборки</h4>
            <div className="summary-grid">
              <div className="summary-item">
                <strong>Среднее закрытие</strong>
                <span>{stats.closeAvg.toFixed(2)}</span>
              </div>
              <div className="summary-item">
                <strong>Объём суммарно</strong>
                <span>{stats.volSum.toFixed(2)}</span>
              </div>
              <div className="summary-item">
                <strong>Изменение цены</strong>
                <span>{stats.priceChange.toFixed(2)}</span>
              </div>
              <div className="summary-item">
                <strong>Волатильность</strong>
                <span>{stats.volatility.toFixed(2)}</span>
              </div>
            </div>
          </div>
        )}
  
        {candles.length > 0 && (
          <div style={{ margin: '1rem 0' }}>
            <button onClick={handleExportCSV}>Скачать CSV</button>
          </div>
        )}
  
        {!loading && candles.length === 0 && query && <p>Нет данных по выбранным параметрам.</p>}
        {candles.length > 0 && (
          <div className="table-wrapper">
            <CandleTable candles={candles} visibleCount={visibleCount} />
          </div>
        )}


        {visibleCount < Math.min(candles.length, query?.total || 0) && (
          <button
            onClick={handleLoadMore}
            disabled={loading}
            className="load-more-button"
          >
          Загрузить ещё
          </button>
        )}

      </div>
  
      <div className="candles-sidebar">
        <CandleForm onSubmit={handleQuery} loading={loading} />
        {error && <p style={{ color: 'red' }}>{error}</p>}
      </div>
    </div>
  );  
}

function deduplicate(candles: Candle[] | undefined): Candle[] {
  if (!candles || candles.length === 0) return [];
  const seen = new Set<string>();
  return candles.filter((c) => {
    const key = `${c.symbol}-${c.open_time.seconds}`;
    if (seen.has(key)) return false;
    seen.add(key);
    return true;
  });
}

export default HistoricalCandlesPage;
