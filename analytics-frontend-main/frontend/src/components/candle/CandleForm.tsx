// src/components/CandleForm.tsx

import React, { useState, useEffect } from 'react';
import { fetchSymbols } from '@api/candles';
import type { Interval } from '../../types/candle';
import './CandleForm.css';

interface CandleFormProps {
  onSubmit: (params: {
    symbol: string;
    interval: Interval;
    start: string;
    end: string;
    total: number;
  }) => void;
  loading: boolean;
}

const intervals: Interval[] = ['1m', '5m', '15m', '1h', '4h', '1d'];

function CandleForm({ onSubmit, loading }: CandleFormProps) {
  const [symbols, setSymbols] = useState<string[]>([]);
  const [symbol, setSymbol] = useState('');
  const [interval, setInterval] = useState<Interval>('1h');
  const [start, setStart] = useState('');
  const [end, setEnd] = useState('');
  const [total, setTotal] = useState(500);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchSymbols()
      .then(setSymbols)
      .catch(() => setError('Не удалось загрузить список символов'));
  }, []);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!symbol || !start || !end || !total) return setError('Заполните все поля');
    if (new Date(start) >= new Date(end)) return setError('Дата начала должна быть меньше даты окончания');
    if (total < 1 || total > 2000) return setError('Количество записей должно быть от 1 до 2000');
    setError(null);
    onSubmit({
      symbol,
      interval,
      start: new Date(start).toISOString(),
      end: new Date(end).toISOString(),
      total,
    });
  };

  return (
    <form onSubmit={handleSubmit} className="candle-form">
      <h3>Фильтрация данных</h3>
      {error && <div className="form-error">{error}</div>}

      <div className="form-group">
        <label className="form-label">Символ:</label>
        <select
          className="form-select"
          value={symbol}
          onChange={(e) => setSymbol(e.target.value)}
          required
        >
          <option value="">-- выберите --</option>
          {symbols.map((s) => (
            <option key={s} value={s}>{s}</option>
          ))}
        </select>
      </div>

      <div className="form-group">
        <label className="form-label">Интервал:</label>
        <select
          className="form-select"
          value={interval}
          onChange={(e) => setInterval(e.target.value as Interval)}
        >
          {intervals.map((intv) => (
            <option key={intv} value={intv}>{intv}</option>
          ))}
        </select>
      </div>

      <div className="form-group">
        <label className="form-label">Начало:</label>
        <input
          className="form-input"
          type="datetime-local"
          value={start}
          onChange={(e) => setStart(e.target.value)}
          required
        />
      </div>

      <div className="form-group">
        <label className="form-label">Конец:</label>
        <input
          className="form-input"
          type="datetime-local"
          value={end}
          onChange={(e) => setEnd(e.target.value)}
          required
        />
      </div>

      <div className="form-group">
        <label className="form-label">Количество свечей (1–2000):</label>
        <input
          className="form-input"
          type="number"
          value={total}
          onChange={(e) => setTotal(parseInt(e.target.value))}
          min={1}
          max={2000}
          required
        />
      </div>

      <button type="submit" disabled={loading}>
        Загрузить
      </button>
    </form>
  );
}

export default CandleForm;
