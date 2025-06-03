// src/pages/OfflineAnalysisPage.tsx
import { useState } from 'react';
import CsvUploader from '@components/analytics_csv/CsvUploader';
import api from '@api/axios';
import AnalysisChart from '@components/analytics_csv/AnalysisChart';

import type { CSVParsedCandle, Candle } from '../../types/candle';
import type { AnalyticsResponse } from '../../types/analytics';

import BodyWickInsight from '@components/analytics_csv/BodyWickInsight';
import MetricCards from '@components/analytics_csv/MetricCards';
import VolumeByHourHistogram from '@components/analytics_csv/VolumeByHourHistogram';
import PieChartUpDown from '@components/analytics_csv/PieChartUpDown';
import MetricCommentary from '@components/analytics_csv/MetricCommentary';

import './OfflineAnalysisPage.css';

function OfflineAnalysisPage() {
  const [candles, setCandles] = useState<CSVParsedCandle[]>([]);
  const [analyzing, setAnalyzing] = useState(false);
  const [result, setResult] = useState<AnalyticsResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  const toProtoCandle = (c: CSVParsedCandle): Candle => ({
    symbol: c.symbol,
    open_time: { seconds: Math.floor(c.open_time.getTime() / 1000) },
    close_time: { seconds: Math.floor(c.close_time.getTime() / 1000) },
    open: c.open,
    high: c.high,
    low: c.low,
    close: c.close,
    volume: c.volume,
  });

  const toProtoCandleFields = (raw: any): AnalyticsResponse['analytics'] => {
    const fix = (c: any): Candle => ({
      symbol: c.symbol,
      open_time: { seconds: Math.floor(new Date(c.open_time).getTime() / 1000) },
      close_time: { seconds: Math.floor(new Date(c.close_time).getTime() / 1000) },
      open: c.open,
      high: c.high,
      low: c.low,
      close: c.close,
      volume: c.volume,
    });
    return {
      ...raw,
      max_candle: fix(raw.max_candle),
      min_candle: fix(raw.min_candle),
      most_volatile_candle: fix(raw.most_volatile_candle),
      most_volume_candle: fix(raw.most_volume_candle),
      max_gap_up_candle: fix(raw.max_gap_up_candle),
      max_gap_down_candle: fix(raw.max_gap_down_candle),
    };
  };

  const handleAnalyze = async () => {
    setAnalyzing(true);
    setError(null);
    try {
      const res = await api.post<AnalyticsResponse>('/analyze-csv', candles);
      const adapted: AnalyticsResponse = {
        ...res.data,
        analytics: toProtoCandleFields(res.data.analytics),
      };
      setResult(adapted);
    } catch {
      setError('Ошибка анализа CSV');
    } finally {
      setAnalyzing(false);
    }
  };

  return (
    <div>
      <h2>Анализ загруженного CSV</h2>
      <CsvUploader
        onParsed={(parsed) => {
          setCandles(parsed);
          setError(null);
        }}
        onError={(msg) => {
          setCandles([]);
          setError(msg);
        }}
      />

      {error && <p style={{ color: 'red' }}>{error}</p>}

      {candles.length > 0 && (
        <>
          <p>Загружено {candles.length} свечей.</p>
          <button onClick={handleAnalyze} disabled={analyzing}>
            {analyzing ? 'Анализируем...' : 'Проанализировать'}
          </button>
        </>
      )}

      {result?.analytics && (
        <div>
          <h3>График свечей</h3>
          <AnalysisChart
            key={result.symbol}
            candles={candles.map(toProtoCandle)}
            analytics={result.analytics}
          />

          <MetricCards analytics={result.analytics} />
          <MetricCommentary analytics={result.analytics} />

          <div className="grid-two-cols">
            <VolumeByHourHistogram candles={candles.map(toProtoCandle)} />
            <PieChartUpDown
              upRatio={result.analytics.up_ratio}
              downRatio={result.analytics.down_ratio}
            />
          </div>
          
          <BodyWickInsight analytics={result.analytics} />
        </div>
      )}
    </div>
  );
}

export default OfflineAnalysisPage;