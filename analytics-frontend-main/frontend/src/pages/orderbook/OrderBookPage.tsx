import { useState } from 'react';
import './OrderBookPage.css';
import OrderBookForm from '@components/orderbook/OrderBookForm';
import OrderBookMetrics from '@components/orderbook/OrderBookMetrics';
import OrderBookSpreadChart from '@components/orderbook/OrderBookSpreadChart';
import OrderBookSpreadExplanation from '@components/orderbook/OrderBookSpreadExplanation';
import OrderBookAnimator from '@components/orderbook/OrderBookAnimator';
import OrderBookSummary from '@components/orderbook/OrderBookSummary';
import { fetchOrderBook } from '@api/orderbook';
import type { OrderBookSnapshot, OrderBookAnalysis } from '../../types/orderbook';
import OrderBookAnimatorExplanation from '@components/orderbook/OrderBookAnimatorExplanation';

function OrderBookPage() {
  const [analysis, setAnalysis] = useState<OrderBookAnalysis | null>(null);
  const [snapshots, setSnapshots] = useState<OrderBookSnapshot[]>([]);
  const [query, setQuery] = useState<{ symbol: string; start: string; end: string; pageSize: number } | null>(null);
  const [pageTokens, setPageTokens] = useState<string[]>(['']);
  const [pageIndex, setPageIndex] = useState(0);
  const [nextToken, setNextToken] = useState<string | undefined>(undefined);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [noData, setNoData] = useState(false);

  const handleQuery = async (params: { symbol: string; start: string; end: string; pageSize: number }) => {
    setError(null);
    setLoading(true);
    setNoData(false);
    setQuery(params);
    setPageTokens(['']);
    setPageIndex(0);
    try {
      const res = await fetchOrderBook({ ...params });
      const snaps = res.snapshots ?? [];
      setAnalysis(res.analysis);
      setSnapshots(snaps);
      setNextToken(res.next_page_token);
      if (snaps.length === 0) setNoData(true);
    } catch {
      setError('Ошибка загрузки данных');
      setSnapshots([]);
      setAnalysis(null);
    } finally {
      setLoading(false);
    }
  };

  const loadPage = async (index: number) => {
    if (!query || index < 0 || index >= pageTokens.length) return;
    setLoading(true);
    try {
      const token = pageTokens[index];
      const res = await fetchOrderBook({ ...query, pageToken: token });
      const snaps = res.snapshots ?? [];
      setSnapshots(snaps);
      setAnalysis(res.analysis);
      setNextToken(res.next_page_token);
      setPageIndex(index);
      setNoData(snaps.length === 0);
    } catch {
      setError('Ошибка при загрузке страницы данных');
    } finally {
      setLoading(false);
    }
  };

  const handleLoadNext = async () => {
    if (!query || !nextToken) return;
    try {
      setLoading(true);
      const res = await fetchOrderBook({ ...query, pageToken: nextToken });
      const snaps = res.snapshots ?? [];
      setSnapshots(snaps);
      setAnalysis(res.analysis);
      setNoData(snaps.length === 0);
  
      const newTokens = [...pageTokens];
      newTokens[pageIndex + 1] = nextToken;
      setPageTokens(newTokens);
      setPageIndex(pageIndex + 1);
      setNextToken(res.next_page_token);
    } catch {
      setError('Ошибка при загрузке следующей страницы');
    } finally {
      setLoading(false);
    }
  };
  

  const handleLoadPrev = async () => {
    if (pageIndex > 0) await loadPage(pageIndex - 1);
  };

  return (
    <div className="orderbook-layout">
      <div className="orderbook-main">
        {error && <p className="error-text">{error}</p>}
        {noData && <p className="info-text">Нет данных в выбранном диапазоне</p>}

        {analysis && (
          <>
            <OrderBookMetrics analysis={analysis} />
            <OrderBookSummary analysis={analysis} />
          </>
        )}

        {snapshots.length > 0 && (
          <>
            <div className="chart-analysis-block">
              <div className="chart-block">
                <OrderBookSpreadChart snapshots={snapshots} />
              </div>
              <div className="explanation-block">
                <OrderBookSpreadExplanation/>
              </div>
            </div>

            <OrderBookAnimator snapshots={snapshots} />
            <OrderBookAnimatorExplanation />
            <div className="orderbook-pagination">
              <button onClick={handleLoadPrev} disabled={loading || pageIndex === 0}>← Предыдущие</button>
              {nextToken && (
                <button onClick={handleLoadNext} disabled={loading}>
                  {loading ? 'Загрузка...' : 'Следующие →'}
                </button>
              )}
            </div>
          </>
        )}
      </div>

      <div className="orderbook-sidebar">
        <OrderBookForm onSubmit={handleQuery} loading={loading} />
      </div>
    </div>
  );
}

export default OrderBookPage;
