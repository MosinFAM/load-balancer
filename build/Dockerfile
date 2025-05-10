FROM golang:1.23

WORKDIR /app

# Устанавливаем зависимости для OpenSSL
RUN apt-get update && apt-get install -y pkg-config libssl-dev

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN go build -o balancer ./cmd/balancer

EXPOSE 8080

CMD ["./balancer", "-config", "./config/config.json"]
