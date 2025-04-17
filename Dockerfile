FROM golang:1.23-alpine

WORKDIR /app

# Установка необходимых утилит, включая postgres-client для pg_isready
RUN apk add --no-cache bash curl postgresql-client

# Настройка прав доступа и кэша
ENV GOCACHE=/tmp/go-cache
ENV GOMODCACHE=/tmp/go-mod-cache
RUN mkdir -p /tmp/go-cache /tmp/go-mod-cache && chmod -R 777 /tmp

# Сначала копируем только файлы go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Теперь копируем остальной код
COPY . .

# Создаем директорию для загруженных файлов с правильными правами
RUN mkdir -p /app/uploads && chmod 777 /app/uploads

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -o migrate ./cmd/migration/main.go

# Указываем порт, который будет использовать приложение
EXPOSE 8080

# Скрипт запуска с миграцией
COPY docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

# Запускаем приложение
ENTRYPOINT ["/docker-entrypoint.sh"]
