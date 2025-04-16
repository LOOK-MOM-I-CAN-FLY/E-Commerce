FROM golang:1.23.0-alpine

WORKDIR /app

COPY . .

# Создаем директорию для загруженных файлов
RUN mkdir -p /app/uploads && chmod 755 /app/uploads

RUN go build -o app ./cmd/main.go

EXPOSE 8080

CMD ["/app/app"]
