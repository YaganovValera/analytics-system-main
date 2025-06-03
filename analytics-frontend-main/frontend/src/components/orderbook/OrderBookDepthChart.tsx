import './OrderBookDepthChart.css';
import type { OrderBookSnapshot } from '../../types/orderbook';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';

interface Props {
  snapshots: OrderBookSnapshot[];
  index: number;
  setIndex: (i: number) => void;
}

function OrderBookDepthChart({ snapshots, index, setIndex }: Props) {
  if (!snapshots || snapshots.length === 0 || !snapshots[index]) return null;

  const snapshot = snapshots[index];
  const bids = [...(snapshot.bids ?? [])].sort((a, b) => b.price - a.price);
  const asks = [...(snapshot.asks ?? [])].sort((a, b) => a.price - b.price);

  let cumulativeBid = 0;
  let cumulativeAsk = 0;
  const data: { price: number; bid: number | null; ask: number | null }[] = [];

  for (const b of bids) {
    cumulativeBid += b.quantity || 0;
    data.push({ price: b.price, bid: +cumulativeBid.toFixed(4), ask: null });
  }
  for (const a of asks) {
    cumulativeAsk += a.quantity || 0;
    data.push({ price: a.price, bid: null, ask: +cumulativeAsk.toFixed(4) });
  }

  const sortedData = data.sort((a, b) => a.price - b.price);
  const timestamp = new Date(snapshot.timestamp.seconds * 1000)
  .toLocaleString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });


  return (
    <div className="depth-chart">
      <div className="depth-chart-header">
        <h3>Глубина стакана</h3>
        <p>Кадр {index + 1} из {snapshots.length} • Время: {timestamp}</p>
      </div>

      <input
        type="range"
        min={0}
        max={snapshots.length - 1}
        value={index}
        onChange={(e) => setIndex(Number(e.target.value))}
        className="depth-slider"
      />

      <ResponsiveContainer width="100%" height={400}>
        <LineChart data={sortedData} margin={{ top: 20, right: 40, left: 10, bottom: 20 }}>
          <CartesianGrid strokeDasharray="3 3" />
          <XAxis dataKey="price" tick={{ fontSize: 10 }} type="number" domain={['auto', 'auto']} />
          <YAxis tick={{ fontSize: 10 }} />
          <Tooltip
            formatter={(val: number) => val.toFixed(4)}
            labelFormatter={(label) => `Цена: ${label}`}
          />
          <Line type="stepAfter" dataKey="bid" stroke="#007bff" strokeWidth={2} dot={false} />
          <Line type="stepAfter" dataKey="ask" stroke="#dc3545" strokeWidth={2} dot={false} />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}

export default OrderBookDepthChart;
