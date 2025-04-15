FROM golang:1.23.0-alpine

WORKDIR /app

COPY . .

RUN go build -o app ./cmd/main.go

CMD ["/app/app"]
