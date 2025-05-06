FROM golang:1.24.2-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o /kode-app ./cmd/kode/main.go

FROM alpine
WORKDIR /app

COPY --from=builder /kode-app .
COPY config/testconfig.yaml ./config/

RUN mv ./config/testconfig.yaml ./config/config.yaml && \
    mkdir "storage"

EXPOSE 8080
EXPOSE 1234

CMD ["./kode-app"]