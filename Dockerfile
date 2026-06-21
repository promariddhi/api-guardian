# ---------- Build Stage ----------
FROM golang:1.24 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -o api-guardian \
    ./cmd/gateway

# ---------- Runtime Stage ----------
FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/api-guardian .
COPY configs/gateway.yaml ./configs/gateway.yaml

EXPOSE 8080

CMD ["./api-guardian"]