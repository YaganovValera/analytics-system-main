// src/components/orderbook/OrderBookForm.tsx
import { useEffect, useState } from 'react';
import { fetchSymbols } from '@api/candles';
import './OrderBookForm.css';

interface OrderBookFormProps {
  onSubmit: (params: {
    symbol: string;
    start: string;
    end: string;
    pageSize: number;
  }) => void;
  loading: boolean;
}

function OrderBookForm({ onSubmit, loading }: OrderBookFormProps) {
  const [symbols, setSymbols] = useState<string[]>([]);
  const [symbol, setSymbol] = useState('');
  const [start, setStart] = useState('');
  const [end, setEnd] = useState('');
  const [pageSize, setPageSize] = useState(100);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchSymbols()
      .then(setSymbols)
      .catch(() => setError('Не удалось загрузить список символов'));
  }, []);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (!symbol || !start || !end || !pageSize) {
      return setError('Пожалуйста, заполните все поля');
    }

    const startTime = new Date(start);
    const endTime = new Date(end);
    if (isNaN(startTime.getTime()) || isNaN(endTime.getTime())) {
      return setError('Неверный формат даты');
    }

    if (startTime >= endTime) {
      return setError('Дата начала должна быть раньше даты окончания');
    }

    if (pageSize < 1 || pageSize > 250) {
      return setError('Количество снимков должно быть от 1 до 250');
    }

    setError(null);

    onSubmit({
      symbol,
      start: startTime.toISOString(),
      end: endTime.toISOString(),
      pageSize,
    });
  };

  return (
    <form className="orderbook-form" onSubmit={handleSubmit}>
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
        <label className="form-label">Начало:</label>
        <input
          type="datetime-local"
          value={start}
          onChange={(e) => setStart(e.target.value)}
          required
        />
      </div>

      <div className="form-group">
        <label className="form-label">Конец:</label>
        <input
          type="datetime-local"
          value={end}
          onChange={(e) => setEnd(e.target.value)}
          required
        />
      </div>

      <div className="form-group">
        <label className="form-label">Количество снимков (1–250):</label>
        <input
          type="number"
          value={pageSize}
          onChange={(e) => setPageSize(parseInt(e.target.value))}
          min={1}
          max={250}
          required
        />
      </div>

      <button type="submit" disabled={loading}>
        {loading ? 'Загрузка...' : 'Загрузить'}
      </button>
    </form>
  );
}

export default OrderBookForm;
