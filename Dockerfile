FROM golang:1.23-alpine AS builder

# Метаданные билда
ARG BUILD_DATE
ARG VCS_REF
ARG VERSION

WORKDIR /app

# Установка необходимых утилит
RUN apk add --no-cache bash curl postgresql-client ca-certificates git

# Настройка прав доступа и кэша
ENV GOCACHE=/tmp/go-cache
ENV GOMODCACHE=/tmp/go-mod-cache
RUN mkdir -p /tmp/go-cache /tmp/go-mod-cache && chmod -R 777 /tmp

# Копируем только файлы go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Теперь копируем остальной код
COPY . .

# Собираем приложения с версией и timestamp
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X main.Version=${VERSION} -X main.BuildTime=${BUILD_DATE}" -o app ./cmd/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o migrate ./cmd/migration/main.go

# Создаем директорию для загруженных файлов
RUN mkdir -p /app/uploads && chmod 777 /app/uploads

# Второй этап - минимальный образ
FROM alpine:3.18

# Метаданные билда
ARG BUILD_DATE
ARG VCS_REF
ARG VERSION

# Labels для контейнера (следуя рекомендациям Open Container Initiative)
LABEL org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.revision="${VCS_REF}" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.title="Digital Marketplace" \
      org.opencontainers.image.description="Digital Marketplace Application" \
      org.opencontainers.image.vendor="Marketplace, Inc."

WORKDIR /app

# Установка необходимых утилит для работы приложения
RUN apk add --no-cache bash curl postgresql-client ca-certificates tzdata && \
    cp /usr/share/zoneinfo/UTC /etc/localtime && \
    echo "UTC" > /etc/timezone

# Копируем бинарные файлы из этапа сборки
COPY --from=builder /app/app /app/app
COPY --from=builder /app/migrate /app/migrate
COPY --from=builder /app/docker-entrypoint.sh /app/docker-entrypoint.sh

# Копируем статические файлы и шаблоны
COPY --from=builder /app/web /app/web

# Создаем директорию для загруженных файлов
COPY --from=builder /app/uploads /app/uploads

# Проверка, что файлы скопированы и имеют правильные разрешения
RUN ls -la /app && chmod +x /app/app /app/migrate /app/docker-entrypoint.sh

# Указываем порт, который будет использовать приложение
EXPOSE 8080

# Проверка работоспособности
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Запускаем приложение
ENTRYPOINT ["/app/docker-entrypoint.sh"]
