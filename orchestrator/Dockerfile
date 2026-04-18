FROM golang:1.24-alpine AS builder
WORKDIR /src

COPY api-gateway ./api-gateway
COPY orchestrator ./orchestrator

WORKDIR /src/orchestrator
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /orchestrator ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=builder /orchestrator /orchestrator
EXPOSE 50051
ENTRYPOINT ["/orchestrator"]
