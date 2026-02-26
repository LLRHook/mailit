# Build stage
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X main.Version=${VERSION:-dev}" -o /mailit ./cmd/mailit

# Runtime stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S mailit && adduser -S mailit -G mailit

COPY --from=builder /mailit /usr/local/bin/mailit
COPY --from=builder /app/db/migrations /migrations
COPY --from=builder /app/config/mailit.example.yaml /etc/mailit/mailit.yaml

USER mailit

EXPOSE 8080 25 587

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8080/healthz || exit 1

ENTRYPOINT ["mailit"]
CMD ["serve", "--config", "/etc/mailit/mailit.yaml"]
