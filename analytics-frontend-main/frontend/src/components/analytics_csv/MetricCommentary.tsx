// src/components/MetricCommentary.tsx
import type { AnalyticsResponse } from '../../types/analytics';
import './MetricCommentary.css';

interface Props {
  analytics: AnalyticsResponse['analytics'];
}

function MetricCommentary({ analytics }: Props) {
  const {
    sum_volume,
    price_change,
    price_range_percent,
    bullish_streak,
    bearish_streak,
    dominant_hour,
    max_gap_up,
    max_gap_down,
    up_ratio,
    down_ratio,
  } = analytics;

  const lines: string[] = [];

  // 🔍 Направление движения
  if (price_change > 0 && up_ratio > 0.55) {
    lines.push('За период наблюдался устойчивый рост: цена увеличилась, и большинство свечей были бычьими. Это указывает на сильное покупательское давление.');
  } else if (price_change < 0 && down_ratio > 0.55) {
    lines.push('Рынок находился под давлением продавцов: снижение цены сопровождалось преобладанием медвежьих свечей.');
  } else {
    lines.push('Явного тренда не выявлено — движение было смешанным, без выраженного доминирования покупателя или продавца.');
  }

  // 🌪 Волатильность и диапазон
  if (price_range_percent > 1.5) {
    lines.push(`Диапазон цен составил ${price_range_percent.toFixed(2)}% — это указывает на высокую волатильность. Рынок мог реагировать на внешние факторы или новости.`);
  } else {
    lines.push(`Диапазон движения (${price_range_percent.toFixed(2)}%) был относительно узким, что говорит о сдержанном поведении участников.`);
  }

  // 📊 Объём
  if (sum_volume > 0) {
    lines.push(`Торговая активность была стабильной. Суммарный объём составил ${sum_volume.toFixed(0)}, что даёт достаточную ликвидность для анализа.`);
  }

  // 📈 Стрики
  if (bullish_streak >= 5 || bearish_streak >= 5) {
    if (bullish_streak >= bearish_streak) {
      lines.push(`Замечена сильная восходящая серия: ${bullish_streak} свечей подряд — это подтверждает устойчивость бычьего тренда.`);
    } else {
      lines.push(`Выделяется продолжительный нисходящий участок: ${bearish_streak} свечей подряд — возможна паническая фаза или тренд вниз.`);
    }
  } else {
    lines.push('Рост и падение чередовались, длительных трендовых серий не наблюдалось.');
  }

  // ⏰ Часы активности
  lines.push(`Наибольшая концентрация объёма пришлась на ${dominant_hour}:00 UTC. Это может быть связано с открытием крупных рынков.`);

  // 🔺 Гэпы
  if (max_gap_up > 0.1 || max_gap_down > 0.1) {
    lines.push(`Зафиксированы значимые гэпы: вверх до ${max_gap_up.toFixed(2)} и вниз до ${max_gap_down.toFixed(2)}. Это часто сопровождается сильными эмоциональными реакциями.`);
  } else {
    lines.push('Гэпы были незначительными — рынок двигался последовательно, без резких скачков.');
  }

  return (
    <div className="metric-commentary-block">
      <h4>🧠 Интерпретация метрик</h4>
      {lines.map((line, idx) => (
        <p key={idx}>{line}</p>
      ))}
    </div>
  );
}

export default MetricCommentary;