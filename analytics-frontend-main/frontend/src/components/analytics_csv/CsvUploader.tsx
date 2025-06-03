// src/components/CsvUploader.tsx
import React from 'react';
import Papa from 'papaparse';
import type { CSVParsedCandle } from '../../types/candle';
import './CsvUploader.css';

interface CsvUploaderProps {
  onParsed: (data: CSVParsedCandle[]) => void;
  onError?: (message: string) => void;
}

function CsvUploader({ onParsed, onError }: CsvUploaderProps) {
  const handleFile = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    Papa.parse(file, {
      header: true,
      skipEmptyLines: true,
      complete: (result) => {
        try {
          if (!Array.isArray(result.data) || result.data.length === 0) {
            onError?.('Файл пустой или не содержит данных.');
            return;
          }

          const candles: CSVParsedCandle[] = (result.data as any[])
            .map((row) => ({
              symbol: row.symbol,
              open_time: new Date(row.open_time),
              close_time: new Date(row.close_time),
              open: parseFloat(row.open),
              high: parseFloat(row.high),
              low: parseFloat(row.low),
              close: parseFloat(row.close),
              volume: parseFloat(row.volume),
            }))
            .filter((c) =>
              c.symbol &&
              !isNaN(c.open_time.getTime()) &&
              !isNaN(c.close_time.getTime()) &&
              !isNaN(c.open) &&
              !isNaN(c.high) &&
              !isNaN(c.low) &&
              !isNaN(c.close) &&
              !isNaN(c.volume)
            );

          if (candles.length === 0) {
            onError?.('Не удалось распознать ни одной валидной свечи.');
            return;
          }

          onParsed(candles);
        } catch {
          onError?.('Ошибка при разборе CSV. Проверьте формат.');
        }
      },
      error: () => {
        onError?.('Ошибка чтения CSV-файла.');
      },
    });
  };

  return (
    <div className="csv-uploader">
      <input type="file" accept=".csv" onChange={handleFile} />
    </div>
  );
}

export default CsvUploader;
