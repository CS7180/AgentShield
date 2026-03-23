FROM golang:1.22-alpine AS builder

WORKDIR /app

# Download dependencies first (layer cache)
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /api-gateway ./cmd/server

# ── Runtime image ──────────────────────────────────────────────────────────────
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /api-gateway ./api-gateway
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

USER nobody

ENTRYPOINT ["./api-gateway"]
