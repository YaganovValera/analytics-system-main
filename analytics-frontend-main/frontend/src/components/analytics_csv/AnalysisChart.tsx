// src/components/AnalysisChart.tsx
import {
  createChart,
  type CandlestickData,
  type UTCTimestamp,
  type IChartApi,
  type SeriesMarkerPosition,
  type SeriesMarkerShape,
} from 'lightweight-charts';
import { useEffect, useRef } from 'react';
import type { Candle } from '../../types/candle';
import type { AnalyticsResponse } from '../../types/analytics';
import './AnalysisChart.css';

interface Props {
  candles: Candle[];
  analytics: AnalyticsResponse['analytics'];
}

function isValidTimestamp(value: any): value is UTCTimestamp {
  return typeof value === 'number' && !isNaN(value) && value > 0;
}

function AnalysisChart({ candles, analytics }: Props) {
  const chartRef = useRef<HTMLDivElement>(null);
  const chart = useRef<IChartApi | null>(null);

  useEffect(() => {
    if (!chartRef.current || candles.length === 0) return;

    chartRef.current.innerHTML = '';

    const instance = createChart(chartRef.current, {
      width: chartRef.current.clientWidth - 100,
      height: 400,
      layout: {
        background: { color: '#fff' },
        textColor: '#000',
      },
      grid: {
        vertLines: { color: '#eee' },
        horzLines: { color: '#eee' },
      },
      timeScale: { timeVisible: true },
      crosshair: { mode: 1 },
    });

    chart.current = instance;

    const series = instance.addCandlestickSeries();

    const data: CandlestickData[] = candles.map((c) => ({
      time: c.open_time.seconds as UTCTimestamp,
      open: c.open,
      high: c.high,
      low: c.low,
      close: c.close,
    }));

    series.setData(data);

    const markers = [
      {
        candle: analytics.most_volume_candle,
        position: 'aboveBar' as SeriesMarkerPosition,
        color: '#7e57c2',
        shape: 'circle' as SeriesMarkerShape,
        text: 'Max Vol',
      },
      {
        candle: analytics.most_volatile_candle,
        position: 'belowBar' as SeriesMarkerPosition,
        color: '#fbc02d',
        shape: 'circle' as SeriesMarkerShape,
        text: 'Volatile',
      },
      {
        candle: analytics.max_gap_up_candle,
        position: 'aboveBar' as SeriesMarkerPosition,
        color: '#2e7d32',
        shape: 'arrowUp' as SeriesMarkerShape,
        text: '🔺',
      },
      {
        candle: analytics.max_gap_down_candle,
        position: 'belowBar' as SeriesMarkerPosition,
        color: '#c62828',
        shape: 'arrowDown' as SeriesMarkerShape,
        text: '🔻',
      },
    ].filter(m => isValidTimestamp(m.candle?.open_time?.seconds)).map(m => ({
      time: m.candle.open_time.seconds as UTCTimestamp,
      position: m.position,
      color: m.color,
      shape: m.shape,
      text: m.text,
    }));

    markers.sort((a, b) => a.time - b.time);
    series.setMarkers(markers);

    return () => instance.remove();
  }, [candles, analytics]);

  return (
    <div className="analysis-container">
      <div className="chart-block">
        <div ref={chartRef} className="chart-wrapper" />
      </div>
      <div className="legend-block">
        <h4>Обозначения</h4>
        <div className="legend-item">
          <span className="icon green">🔺</span>
          <span className="label">Max Gap Up</span>
          <span className="tooltip" title="Резкий скачок вверх после открытия. Может быть вызван новостями или ордерами.">❓</span>
        </div>
        <div className="legend-item">
          <span className="icon red">🔻</span>
          <span className="label">Max Gap Down</span>
          <span className="tooltip" title="Резкое падение после открытия. Часто свидетельствует о панике на рынке.">❓</span>
        </div>
        <div className="legend-item">
          <span className="icon purple">🟣</span>
          <span className="label">Макс. объём</span>
          <span className="tooltip" title="Свеча с наибольшим объёмом. Указывает на аномальную активность.">❓</span>
        </div>
        <div className="legend-item">
          <span className="icon yellow">🟡</span>
          <span className="label">Макс. волатильность</span>
          <span className="tooltip" title="Свеча с наибольшим разбросом между high и low. Признак нестабильности.">❓</span>
        </div>
      </div>
    </div>
  );
}

export default AnalysisChart;