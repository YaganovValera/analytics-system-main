// src/components/VolumeByHourHistogram.tsx
import { useMemo } from 'react';
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, Cell } from 'recharts';
import type { Candle } from '../../types/candle';
import './VolumeByHourHistogram.css';

interface Props {
  candles: Candle[];
}

function VolumeByHourHistogram({ candles }: Props) {
  const data = useMemo(() => {
    const hourVolume: Record<number, number> = {};
    for (let i = 0; i < 24; i++) hourVolume[i] = 0;

    for (const c of candles) {
      const hour = new Date(c.open_time.seconds * 1000).getUTCHours();
      hourVolume[hour] += c.volume;
    }

    return Object.entries(hourVolume).map(([hour, vol]) => ({
      hour,
      volume: Number(vol.toFixed(2)),
    }));
  }, [candles]);

  return (
    <div className="volume-histogram-block">
      <h3>📊 Объём торгов по часам (UTC)</h3>
      <ResponsiveContainer width="100%" height={300}>
        <BarChart data={data} margin={{ top: 20, right: 30, left: 10, bottom: 10 }}>
          <XAxis dataKey="hour" label={{ value: 'Час (UTC)', position: 'insideBottom', dy: 10 }} />
          <YAxis tickFormatter={(v) => v.toFixed(0)} />
          <Tooltip formatter={(v: number) => v.toFixed(2)} labelFormatter={(l) => `${l}:00`} />
          <Bar dataKey="volume">
            {data.map((_, idx) => (
              <Cell key={`cell-${idx}`} fill="#4caf50" />
            ))}
          </Bar>
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}

export default VolumeByHourHistogram;