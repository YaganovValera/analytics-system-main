// src/components/MetricCards.tsx
import type { AnalyticsResponse } from '../../types/analytics';
import './MetricCards.css';

interface Props {
  analytics: AnalyticsResponse['analytics'];
}

const METRICS = [
  { key: 'avg_close', label: 'Среднее закрытие', icon: '💰', format: (v: number) => v.toFixed(2) },
  { key: 'sum_volume', label: 'Суммарный объём', icon: '📊', format: (v: number) => v.toFixed(2) },
  { key: 'price_change', label: 'Изменение цены', icon: '📈', format: (v: number) => v.toFixed(2) },
  { key: 'volatility', label: 'Волатильность', icon: '🌪️', format: (v: number) => v.toFixed(2) },
  { key: 'price_range_percent', label: 'Диапазон (%)', icon: '📐', format: (v: number) => v.toFixed(2) + '%' },
  { key: 'bullish_streak', label: 'Макс. рост подряд', icon: '📗', format: (v: number) => v },
  { key: 'bearish_streak', label: 'Макс. падение подряд', icon: '📕', format: (v: number) => v },
  { key: 'dominant_hour', label: 'Час пик (UTC)', icon: '⏰', format: (v: number) => v + ':00' },
  { key: 'max_gap_up', label: 'Макс. гэп ↑', icon: '🔺', format: (v: number) => v.toFixed(2) },
  { key: 'max_gap_down', label: 'Макс. гэп ↓', icon: '🔻', format: (v: number) => v.toFixed(2) },
];

function MetricCards({ analytics }: Props) {
  return (
    <div className="metric-cards">
      {METRICS.map(({ key, label, icon, format }) => (
        <div className="metric-card" key={key} title={label}>
          <div className="metric-icon">{icon}</div>
          <div className="metric-value">{format((analytics as any)[key])}</div>
          <div className="metric-label">{label}</div>
        </div>
      ))}
    </div>
  );
}

export default MetricCards;