import type { OrderBookAnalysis } from '../../types/orderbook';
import './OrderBookSummary.css';

interface Props {
  analysis: OrderBookAnalysis;
}

function OrderBookSummary({ analysis }: Props) {
  const {
    avg_spread_percent,
    avg_bid_volume_top10,
    avg_ask_volume_top10,
    imbalance_end,
    bid_slope,
    ask_slope,
    max_bid_wall_price,
    max_bid_wall_volume,
    max_ask_wall_price,
    max_ask_wall_volume,
  } = analysis;

  const spread = avg_spread_percent * 100;
  const imbalance = imbalance_end;

  let spreadText = 'Спред между лучшими заявками находится в норме, что указывает на среднюю ликвидность.';
  if (spread < 0.03) {
    spreadText = 'Очень узкий спред — рынок обладает высокой ликвидностью, сделки исполняются с минимальным проскальзыванием.';
  } else if (spread > 0.2) {
    spreadText = 'Широкий спред — низкая ликвидность. При попытке совершить сделку с объёмом возможны значительные проскальзывания.';
  }

  let imbalanceText = 'Баланс между BID и ASK объёмами в пределах нормы — нет явного перекоса.';
  if (imbalance > 0.3) {
    imbalanceText = 'Наблюдается преобладание покупателей: в BID-сегменте стакана сосредоточено больше ликвидности.';
  } else if (imbalance < -0.3) {
    imbalanceText = 'Преобладание продавцов — объёмы на стороне ASK значительно превышают BID.';
  }

  const volumeText = avg_bid_volume_top10 > avg_ask_volume_top10
    ? 'Покупатели доминируют в ближайших уровнях: BID заявок больше, чем ASK.'
    : avg_ask_volume_top10 > avg_bid_volume_top10
    ? 'В ASK-сегменте наблюдается большая ликвидность — продавцы более активны.'
    : 'BID и ASK объемы сбалансированы — спрос и предложение распределены равномерно.';

  const slopeText = (bid_slope < -1 || ask_slope < -1)
    ? 'Глубина стакана резко убывает вдаль от текущей цены — объёмы сконцентрированы вблизи лучшей цены. Это может привести к волатильности при резких ордерах.'
    : 'Стакан распределён равномерно: ликвидность присутствует как у ближайших уровней, так и дальше от цены. Это снижает риск резких ценовых движений.';

  const wallsText = (max_bid_wall_volume > 0 || max_ask_wall_volume > 0)
    ? `Обнаружены крупные заявки (ценовые "стены"): BID на уровне ${max_bid_wall_price}, объём ${max_bid_wall_volume}; ASK на уровне ${max_ask_wall_price}, объём ${max_ask_wall_volume}. Эти уровни могут служить локальными сопротивлениями и поддержками.`
    : 'Явно выраженных ценовых "стен" в стакане не зафиксировано — структура равномерна.';

  return (
    <div className="orderbook-summary">
      <h3>📊 Интерпретация рыночной структуры</h3>
      <p><strong>Спред:</strong> {spreadText}</p>
      <p><strong>Дисбаланс:</strong> {imbalanceText}</p>
      <p><strong>Объёмы BID vs ASK:</strong> {volumeText}</p>
      <p><strong>Глубина стакана:</strong> {slopeText}</p>
      <p><strong>Крупные заявки:</strong> {wallsText}</p>
    </div>
  );
}

export default OrderBookSummary;
