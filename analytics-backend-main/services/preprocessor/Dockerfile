# services/preprocessor/Dockerfile
# syntax=docker/dockerfile:1

# ====== BUILD STAGE ======
FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# 1) Copy workspace
COPY go.work go.work.sum ./

# 2) Copy go.mod/go.sum of all modules
COPY common/go.mod common/go.sum ./common/
COPY proto/go.mod proto/go.sum ./proto/
COPY services/query-service/go.mod services/query-service/go.sum ./services/query-service/
COPY services/analytics-api/go.mod services/analytics-api/go.sum ./services/analytics-api/
COPY services/api-gateway/go.mod services/api-gateway/go.sum ./services/api-gateway/
COPY services/auth/go.mod services/auth/go.sum ./services/auth/
COPY services/preprocessor/go.mod services/preprocessor/go.sum ./services/preprocessor/
COPY services/market-data-collector/go.mod services/market-data-collector/go.sum ./services/market-data-collector/

# 3) Sync dependencies
RUN go work sync

# 4) Copy source
COPY . .

WORKDIR /app/services/preprocessor

# 5) Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/preprocessor ./cmd/preprocessor

# ====== RUNTIME STAGE ======
FROM alpine:3.18 AS runner

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/preprocessor .
COPY --from=builder /app/services/preprocessor/config ./config
COPY --from=builder /app/services/preprocessor/migrations/timescaledb ./migrations/timescaledb

# Drop privileges
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

EXPOSE 8081

# Optional: add healthcheck
HEALTHCHECK --interval=10s --timeout=3s CMD wget -qO- http://localhost:8081/healthz || exit 1

ENTRYPOINT ["./preprocessor", "--config", "config/config.yaml"]
