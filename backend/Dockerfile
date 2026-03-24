# ── Stage 1: сборка ───────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/app ./cmd/app/...

# ── Stage 2: финальный образ ──────────────────────────────────────
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/bin/app ./app
COPY --from=builder /app/migrations ./migrations

EXPOSE 8081

# Для отладки можно раскомментировать следующую строку, чтобы контейнер не завершался сразу после запуска
#ENTRYPOINT ["sleep", "infinity"]

# Запуск приложения (при отладке, закооментируйте эту строку и раскомментируйте предыдущую)
ENTRYPOINT ["./app"]