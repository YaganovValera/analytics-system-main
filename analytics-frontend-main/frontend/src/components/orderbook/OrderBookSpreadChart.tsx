import './OrderBookSpreadChart.css';
import type { OrderBookSnapshot } from '../../types/orderbook';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
  Brush,
  ReferenceDot,
} from 'recharts';

interface Props {
  snapshots: OrderBookSnapshot[];
}

function OrderBookSpreadChart({ snapshots }: Props) {
  const data = snapshots.map((s) => {
    const time = new Date(s.timestamp.seconds * 1000).toLocaleTimeString();
    const bestBid = s.bids?.[0]?.price ?? 0;
    const bestAsk = s.asks?.[0]?.price ?? 0;
    const spread = bestAsk > 0 && bestBid > 0 ? (bestAsk - bestBid) / bestAsk : 0;

    const bidVolume = s.bids?.slice(0, 10).reduce((acc, lvl) => acc + (lvl.quantity || 0), 0) ?? 0;
    const askVolume = s.asks?.slice(0, 10).reduce((acc, lvl) => acc + (lvl.quantity || 0), 0) ?? 0;
    const imbalance = bidVolume + askVolume > 0
      ? (bidVolume - askVolume) / (bidVolume + askVolume)
      : 0;

    return { time, spread: +(spread * 100).toFixed(4), imbalance: +imbalance.toFixed(4) };
  });

  const maxSpread = data.reduce((max, d) => (d.spread > max.spread ? d : max), data[0]);
  const minImbalance = data.reduce((min, d) => (d.imbalance < min.imbalance ? d : min), data[0]);

  return (
    <div className="spread-chart">
      <div className="spread-chart-header">
        <h3>Динамика спреда и дисбаланса</h3>
        <p>Обрабатывается {snapshots.length} снимков стакана</p>
      </div>
      <ResponsiveContainer width="100%" height={420}>
        <LineChart data={data} margin={{ top: 20, right: 30, left: 10, bottom: 20 }}>
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis dataKey="time" tick={{ fontSize: 10 }} angle={-35} textAnchor="end" height={50} />
          <YAxis yAxisId="left" domain={[0, 'auto']} tickFormatter={(v) => v + '%'} />
          <YAxis yAxisId="right" orientation="right" domain={[-1, 1]} />

          <Tooltip
            formatter={(value: number, name: string) => [value.toFixed(4), name === 'spread' ? 'Спред (%)' : 'Дисбаланс']}
            labelFormatter={(label) => `Время: ${label}`}
          />

          <Line
            yAxisId="left"
            type="monotone"
            dataKey="spread"
            stroke="#005eff"
            strokeWidth={2}
            dot={false}
          />
          <Line
            yAxisId="right"
            type="monotone"
            dataKey="imbalance"
            stroke="#eb4034"
            strokeWidth={2}
            dot={false}
          />

          <ReferenceDot
            x={maxSpread.time}
            y={maxSpread.spread}
            yAxisId="left"
            r={6}
            fill="#005eff"
            stroke="#0033cc"
            strokeWidth={1.5}
            label={{ value: 'макс. спред', position: 'top', fontSize: 10 }}
          />

          <ReferenceDot
            x={minImbalance.time}
            y={minImbalance.imbalance}
            yAxisId="right"
            r={6}
            fill="#eb4034"
            stroke="#a11c1c"
            strokeWidth={1.5}
            label={{ value: 'мин. дисбаланс', position: 'bottom', fontSize: 10 }}
          />

          <Brush dataKey="time" height={20} stroke="#8884d8" />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}

export default OrderBookSpreadChart;
