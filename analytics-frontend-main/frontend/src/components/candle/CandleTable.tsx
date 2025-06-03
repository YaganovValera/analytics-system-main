// src/components/CandleTable.tsx

import { useMemo, useState } from 'react';
import type { Candle } from '../../types/candle';
import './CandleTable.css';

interface CandleTableProps {
  candles: Candle[];
  visibleCount: number;
}

const HEADERS = [
  { key: 'id_candle', label: 'ID' },
  { key: 'open_time', label: 'Открытие' },
  { key: 'close_time', label: 'Закрытие' },
  { key: 'open', label: 'Open' },
  { key: 'high', label: 'High' },
  { key: 'low', label: 'Low' },
  { key: 'close', label: 'Close' },
  { key: 'volume', label: 'Объём' },
];

function CandleTable({ candles, visibleCount }: CandleTableProps) {
  const [sortKey, setSortKey] = useState<string>('open_time');
  const [sortAsc, setSortAsc] = useState<boolean>(true);

  const maxHigh = useMemo(() => Math.max(...candles.map(c => c.high)), [candles]);
  const minLow = useMemo(() => Math.min(...candles.map(c => c.low)), [candles]);

  const sorted = useMemo(() => {
    const copy = [...candles];
    copy.sort((a, b) => {
      const aVal = (a as any)[sortKey];
      const bVal = (b as any)[sortKey];

      if (sortKey === 'open_time' || sortKey === 'close_time') {
        return sortAsc
          ? aVal.seconds - bVal.seconds
          : bVal.seconds - aVal.seconds;
      }
      return sortAsc
        ? aVal > bVal ? 1 : -1
        : aVal < bVal ? 1 : -1;
    });
    return copy;
  }, [candles, sortKey, sortAsc]);

  const visible = sorted.slice(0, visibleCount);

  const handleSort = (key: string) => {
    if (key === sortKey) setSortAsc(!sortAsc);
    else {
      setSortKey(key);
      setSortAsc(true);
    }
  };

  const formatTime = (ts: { seconds: number }) => {
    return new Date(ts.seconds * 1000).toLocaleString();
  };

  return (
    <div className="candle-table-wrapper">
      <h3>Свечи ({candles.length})</h3>
      <table className="candle-table">
        <thead>
          <tr>
            {HEADERS.map(({ key, label }) => (
              <th key={key} onClick={() => handleSort(key)} style={{ cursor: 'pointer' }}>
                {label} {sortKey === key ? (sortAsc ? '▲' : '▼') : ''}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {visible.map((candle, idx) => {
            let rowClass =
              candle.close > candle.open ? 'row-up' : candle.close < candle.open ? 'row-down' : 'row-neutral';

            if (candle.high === maxHigh) rowClass += ' row-max';
            if (candle.low === minLow) rowClass += ' row-min';

            return (
              <tr key={`${candle.symbol}-${candle.open_time.seconds}-${idx}`} className={rowClass}>
                <td>{idx + 1}</td>
                <td>{formatTime(candle.open_time)}</td>
                <td>{formatTime(candle.close_time)}</td>
                <td>{candle.open.toFixed(2)}</td>
                <td>{candle.high.toFixed(2)}</td>
                <td>{candle.low.toFixed(2)}</td>
                <td>{candle.close.toFixed(2)}</td>
                <td>{candle.volume.toFixed(2)}</td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}


export default CandleTable;
