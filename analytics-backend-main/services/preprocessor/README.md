
# 🧠 preprocessor

**`preprocessor`** — это микросервис в системе аналитики криптовалют, отвечающий за агрегацию рыночных событий типа `trade` в OHLCV-свечи, сохранение снапшотов `depthUpdate`, и публикацию готовых данных downstream-сервисам. Он выступает как буфер, форматизатор и маршрутизатор с полной поддержкой восстановления и наблюдаемости.

---

## 🚀 Основные функции

- 🕯️ Агрегация `@trade` событий в свечи по интервалам (`1m`, `5m`, `15m`, ...).
- 📤 Публикация завершённых свечей в Kafka (топики `candles.<interval>`).
- 💾 Запись снапшотов стакана заявок в TimescaleDB (`orderbook_snapshots`).
- ♻️ Восстановление in-progress свечей из Redis при рестарте.
- 📊 Метрики в Prometheus (`/metrics`).
- ❤️ Readiness / Liveness пробы для Kubernetes (`/readyz`, `/healthz`).
- 📁 Миграции TimescaleDB автоматически применяются при запуске.

---

## 🧩 Архитектура

```text
     Kafka (Binance stream)
       ┌────────────┐
       │            ▼
[marketdata.raw]  [marketdata.orderbook]
       │            │
       ▼            ▼
  [ preprocessor microservice ]
         │          │
    ┌────┴───┐  ┌───┴────┐
    │Aggregator│ │Orderbook│
    └────┬────┘  └────┬───┘
         ▼           ▼
  [candles.*]    [TimescaleDB]
      Kafka        ohlcv, book
```

---

## ⚙️ Конфигурация

Настройки берутся из `config/config.yaml`, переопределяются переменными `PREPROCESSOR_`.

Пример:

```yaml
kafka_consumer:
  brokers: ["kafka:9092"]
  group_id: "preprocessor"

kafka_producer:
  brokers: ["kafka:9092"]
  required_acks: "all"

raw_topic: "marketdata.raw"
output_topic_prefix: "candles"
orderbook_topic: "marketdata.orderbook"

redis:
  addr: "redis:6379"

timescaledb:
  dsn: "postgres://user:pass@timescaledb:5432/analytics?sslmode=disable"
  migrations_dir: "/app/migrations/timescaledb"

intervals: ["1m", "5m", "15m", "1h", "1d"]
```

---

## 🧪 Метрики (Prometheus)

Endpoint: `http://localhost:8081/metrics`

| Метрика                          | Назначение                              |
|----------------------------------|------------------------------------------|
| `aggregator_processed_total`     | Обработанные трейды по интервалам        |
| `aggregator_flushed_total`       | Сброшенные свечи                         |
| `aggregator_flush_latency`       | Задержка между последним тиком и сбросом |
| `redis_restore_success_total`    | Успешные восстановления in-progress bar  |
| `kafka_published_total`          | Опубликованные свечи                     |
| `orderbook_processed_total`      | Сохранённые снапшоты стакана заявок      |

---

## 🛡️ Kubernetes

Манифесты:

```
/deploy/k8s/preprocessor/
```

Применение:

```bash
kubectl apply -k deploy/k8s/preprocessor
```

Компоненты:
- `deployment.yaml` — основной Pod.
- `service.yaml` — доступ к порту 8081.
- `configmap.yaml` — параметры приложения.
- `livenessProbe` / `readinessProbe`.

---

## 🐳 Docker

```bash
docker build -t preprocessor -f Dockerfile .
docker run --rm -p 8081:8081 preprocessor --config config/config.yaml
```

Или через Compose:

```bash
docker compose up --build -d preprocessor
```

---

## 📁 Структура проекта

```
├── cmd/preprocessor             # Точка входа
├── config/config.yaml           # Конфигурация
├── internal/
│   ├── app                      # Запуск и управление
│   ├── aggregator               # Candle менеджер
│   ├── kafka                    # Consumers
│   ├── storage/
│   │   ├── timescaledb          # Вставка свечей и стакана
│   │   ├── redis                # Хранилище in-progress баров
│   │   └── kafkasink            # Продюсер в Kafka
│   ├── transport                # Marshaling protobuf/SQL
│   └── metrics                  # Прометеус метрики
└── migrations/timescaledb      # SQL-миграции
```

---

## 📌 Зависимости

- Kafka
- Redis
- TimescaleDB (через pgx)
- Prometheus
- Protobuf (marketdata + analytics)

---

## 🧠 Почему это важно

Этот микросервис:
- обеспечивает надёжную агрегацию огромного потока тиков в компактные свечи;
- гарантирует сохранность данных при сбоях (через Redis);
- централизует публикацию в Kafka для последующего анализа;
- критически важен для таймсерийной аналитики и построения графиков.

---

## 👨‍💻 Отладка

- Kafka:
  ```bash
  kafka-console-consumer.sh --topic marketdata.raw --from-beginning --bootstrap-server kafka:9092
  ```
- Метрики:
  ```bash
  curl http://localhost:8081/metrics
  ```
- Проверка свечей:
  ```sql
  SELECT * FROM candles ORDER BY time DESC LIMIT 10;
  ```

---