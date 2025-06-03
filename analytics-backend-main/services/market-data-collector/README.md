
# market-data-collector

📡 **`market-data-collector`** — это микросервис системы аналитики криптовалют, предназначенный для сбора рыночных данных в реальном времени с Binance WebSocket API и их публикации в Apache Kafka. Он обеспечивает надёжную маршрутизацию событий типа `trade` и `depthUpdate` с минимальной задержкой и высокой наблюдаемостью.

---

## 🚀 Основные функции

- 🔌 Подключение к Binance WebSocket (с backoff, ping, reconnect).
- 📦 Поддержка потоков `trade` и `depthUpdate` по множеству символов.
- 📤 Публикация событий в Kafka в сериализованном формате Protobuf.
- 📊 Метрики в формате Prometheus (`/metrics`).
- ❤️ Health/readiness пробы для Kubernetes.
- 📈 Гибкая буферизация (flush, batching) и сжатие Kafka сообщений.

---

## 🧩 Архитектура

```text
           Binance WS
               ↓
     [market-data-collector]
               ↓
    ┌──────────────┬──────────────┐
    │ trade stream │ depth stream│
    └──────────────┴──────────────┘
               ↓
         Apache Kafka
               ↓
      (другие микросервисы)
````

---

## ⚙️ Конфигурация

Все настройки берутся из `config/config.yaml`, но могут быть переопределены через переменные окружения с префиксом `MARKETDATA_`.

### Примеры ключевых настроек:

```yaml
binance:
  ws_url: "wss://stream.binance.com:9443/ws"
  streams:
    - "btcusdt@trade"
    - "btcusdt@depth"
  read_timeout: "30s"

kafka:
  brokers: ["kafka:9092"]
  raw_topic: "marketdata.raw"
  orderbook_topic: "marketdata.orderbook"
  flush_frequency: "100ms"
  flush_messages: 100
  compression: "gzip"
```

---

## 🧪 Метрики

Экспонируются по адресу: `http://localhost:8086/metrics`

| Метрика                             | Назначение                       |
| ----------------------------------- | -------------------------------- |
| `processor_events_total`            | Общее количество событий         |
| `processor_parse_errors_total`      | Ошибки при парсинге JSON         |
| `processor_serialize_errors_total`  | Ошибки при сериализации protobuf |
| `processor_publish_errors_total`    | Ошибки публикации в Kafka        |
| `processor_publish_latency_seconds` | Задержка публикации              |

---

## 🛡️ Kubernetes

Все манифесты Kubernetes находятся в:

```
/deploy/k8s/market-data-collector/
```

Применение:

```bash
kubectl apply -k deploy/k8s/market-data-collector
```

Компоненты:

* `deployment.yaml` — основное описание Pod и контейнера.
* `service.yaml` — кластерный сервис (порт `8086`).
* `configmap.yaml` — конфигурация приложения (`config.yaml`).
* `livenessProbe`, `readinessProbe` — `/healthz`, `/readyz`.

---

## 🐳 Локальный запуск (через Docker)

```bash
docker build -t market-data-collector -f Dockerfile .
docker run --rm -p 8086:8086 market-data-collector --config config/config.yaml
```

---

## 📌 Зависимости

* Binance WebSocket API
* Apache Kafka
* Prometheus (для сбора метрик)
* TimescaleDB (на следующем этапе — downstream consumer)
* Protobuf (для сериализации сообщений)

---

## 📁 Структура проекта

```
├── cmd/collector          # Точка входа
├── config/config.yaml     # Конфигурация приложения
├── internal/              # Бизнес-логика
│   ├── app                # Run(...) orchestrator
│   ├── processor          # trade / depth обработчики
│   └── transport/binance  # ws-manager, client
├── pkg/binance            # WS-connector (low-level)
├── Dockerfile             # Multi-stage build
└── README.md              # Документация
```

---

## 🧠 Почему он важен

Этот микросервис — **точка входа всей системы рыночной аналитики**. Он гарантирует:

* корректную запись "сырых" данных с биржи,
* минимальные потери при reconnect,
* соблюдение формата и протоколов (Kafka + Protobuf),
* стабильность и масштабируемость (через Kubernetes).

---

## 👨‍💻 Поддержка

В случае проблем:

* Проверь лог `kubectl logs market-data-collector-XYZ`
* Убедись, что Kafka доступен (`kafka:9092`)
* Проверь `/metrics` и `/healthz` на `:8086`

---


