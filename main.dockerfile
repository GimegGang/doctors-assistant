FROM golang:1.24.2-alpine AS builder

# Устанавливаем зависимости компилятора в отдельном слое
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o /kode-app ./cmd/kode/main.go

FROM alpine
WORKDIR /app

COPY --from=builder /kode-app .
COPY config/config.yaml ./config/

RUN mkdir "storage"

EXPOSE 8080
CMD ["./kode-app"]