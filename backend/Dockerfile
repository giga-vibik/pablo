FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o service ./cmd/service

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/service ./service

# config.yaml намеренно НЕ копируется в образ: в нём ключи, а правка настроек
# не должна требовать пересборки. Он монтируется из docker-compose.

EXPOSE 8080

ENTRYPOINT ["./service"]
