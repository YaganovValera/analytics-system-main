import './OrderBookLevelTable.css';
import type { OrderBookSnapshot } from '../../types/orderbook';

interface Props {
  snapshot: OrderBookSnapshot;
}

function OrderBookLevelTable({ snapshot }: Props) {
  const bids = [...(snapshot.bids ?? [])].sort((a, b) => b.price - a.price);
  const asks = [...(snapshot.asks ?? [])].sort((a, b) => a.price - b.price);
  const maxRows = Math.max(bids.length, asks.length);

  return (
    <div className="orderbook-table">
      <div className="orderbook-table-header">
        <h3>Таблица заявок </h3>
      </div>
      <table>
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
  );
}

export default OrderBookLevelTable;
