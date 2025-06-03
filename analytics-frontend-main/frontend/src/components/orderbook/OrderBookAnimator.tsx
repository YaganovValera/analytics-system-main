import './OrderBookAnimator.css';
import { useState } from 'react';
import type { OrderBookSnapshot } from '../../types/orderbook';
import OrderBookDepthChart from './OrderBookDepthChart';

interface Props {
  snapshots: OrderBookSnapshot[];
}

function OrderBookAnimator({ snapshots }: Props) {
  const [index, setIndex] = useState(0);
  const total = snapshots.length;
  const snapshot = snapshots[index];

  const bids = [...(snapshot.bids ?? [])].sort((a, b) => b.price - a.price);
  const asks = [...(snapshot.asks ?? [])].sort((a, b) => a.price - b.price);
  const maxRows = Math.max(bids.length, asks.length);

  const handlePrev = () => {
    if (index > 0) setIndex(index - 1);
  };

  const handleNext = () => {
    if (index < total - 1) setIndex(index + 1);
  };

  return (
    <div className="orderbook-animator">
      <div className="orderbook-animator-header">
        <h3>Анимация стакана заявок</h3>
        <p>Кадр {index + 1} из {total}</p>
      </div>

      <div className="orderbook-animator-controls">
        <button onClick={handlePrev} disabled={index === 0}>← Назад</button>
        <button onClick={handleNext} disabled={index === total - 1}>Вперёд →</button>
      </div>

      <OrderBookDepthChart
        snapshots={snapshots}
        index={index}
        setIndex={setIndex}
      />

      <div className="orderbook-animator-table-wrapper">
        <div className="orderbook-animator-table-header">
          <h4>📋 Таблица заявок (snapshot #{index + 1})</h4>
        </div>
        <div className="orderbook-animator-table-scroll">
          <table className="orderbook-table">
            <thead>
              <tr>
                <th>BID Объём</th>
                <th>BID Цена</th>
                <th>ASK Цена</th>
                <th>ASK Объём</th>
              </tr>
            </thead>
            <tbody>
              {Array.from({ length: maxRows }).map((_, i) => (
                <tr key={i}>
                  <td className="bid">{bids[i]?.quantity?.toFixed(4) ?? ''}</td>
                  <td className="bid">{bids[i]?.price?.toFixed(2) ?? ''}</td>
                  <td className="ask">{asks[i]?.price?.toFixed(2) ?? ''}</td>
                  <td className="ask">{asks[i]?.quantity?.toFixed(4) ?? ''}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

export default OrderBookAnimator;
