import './OrderBookMetrics.css';
import { useState } from 'react';
import type { OrderBookAnalysis } from '../../types/orderbook';

interface OrderBookMetricsProps {
  analysis: OrderBookAnalysis;
}

const formatPercent = (value: number) => (value * 100).toFixed(4) + '%';
const formatFloat = (value: number) => value.toFixed(4);

const metricDescriptions: Record<string, string> = {
  spread: 'Разница между лучшим ASK и BID. Чем меньше, тем выше ликвидность.',
  avg_bid: 'Средний суммарный объём заявок на покупку в топ-10 ценах.',
  avg_ask: 'Средний суммарный объём заявок на продажу в топ-10 ценах.',
  imbalance_start: 'Начальный дисбаланс между BID и ASK. Показывает перекос рынка.',
  imbalance_end: 'Конечный дисбаланс. Сравни с началом для оценки динамики.',
  bid_slope: 'Насколько быстро уменьшается глубина BID. Острый наклон = слабая поддержка.',
  ask_slope: 'То же для ASK. Острый наклон = слабое сопротивление.',
  bid_wall: 'Самая крупная заявка на покупку: объём и цена.',
  ask_wall: 'Самая крупная заявка на продажу: объём и цена.',
};

function FlipCard({
  label,
  value,
  explanation,
}: {
  label: string;
  value: string;
  explanation: string;
}) {
  const [flipped, setFlipped] = useState(false);

  return (
    <div className="orderbook-metric-card" onClick={() => setFlipped(!flipped)}>
      {!flipped ? (
        <div className="orderbook-card-face orderbook-card-front">
          <h4>{label}</h4>
          <p>{value}</p>
        </div>
      ) : (
        <div className="orderbook-card-face orderbook-card-back">
          <p>{explanation}</p>
        </div>
      )}
    </div>
  );
}

function OrderBookMetrics({ analysis }: OrderBookMetricsProps) {
  return (
    <div className="orderbook-metrics-grid">
      <FlipCard label="Спред (%)" value={formatPercent(analysis.avg_spread_percent)} explanation={metricDescriptions.spread} />
      <FlipCard label="BID объём (топ 10)" value={formatFloat(analysis.avg_bid_volume_top10)} explanation={metricDescriptions.avg_bid} />
      <FlipCard label="ASK объём (топ 10)" value={formatFloat(analysis.avg_ask_volume_top10)} explanation={metricDescriptions.avg_ask} />
      <FlipCard label="Дисбаланс (начало)" value={formatFloat(analysis.imbalance_start)} explanation={metricDescriptions.imbalance_start} />
      <FlipCard label="Дисбаланс (конец)" value={formatFloat(analysis.imbalance_end)} explanation={metricDescriptions.imbalance_end} />
      <FlipCard label="Наклон BID" value={formatFloat(analysis.bid_slope)} explanation={metricDescriptions.bid_slope} />
      <FlipCard label="Наклон ASK" value={formatFloat(analysis.ask_slope)} explanation={metricDescriptions.ask_slope} />
      <FlipCard label="BID стена" value={`${analysis.max_bid_wall_volume} @ ${analysis.max_bid_wall_price}`} explanation={metricDescriptions.bid_wall} />
      <FlipCard label="ASK стена" value={`${analysis.max_ask_wall_volume} @ ${analysis.max_ask_wall_price}`} explanation={metricDescriptions.ask_wall} />
    </div>
  );
}

export default OrderBookMetrics;
